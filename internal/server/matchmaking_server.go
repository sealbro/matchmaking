package server

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log/slog"
	gen "matchmaking/generated/grpc"
	"matchmaking/internal/matchmaking"
	"matchmaking/internal/metrics"
	"sync"
)

type MatchmakingServer struct {
	gen.UnimplementedMatchmakingServer
	logger       *slog.Logger
	service      *matchmaking.Service
	playerStates map[string]grpc.ServerStreamingServer[gen.StatusResponse]
	l            sync.RWMutex
}

func NewMatchmakingServer(logger *slog.Logger, service *matchmaking.Service) *MatchmakingServer {
	return &MatchmakingServer{
		logger:       logger,
		service:      service,
		playerStates: make(map[string]grpc.ServerStreamingServer[gen.StatusResponse], 10),
	}
}

func (s *MatchmakingServer) Register() func(registrar grpc.ServiceRegistrar) {
	return func(registrar grpc.ServiceRegistrar) {
		gen.RegisterMatchmakingServer(registrar, s)
	}
}

func (s *MatchmakingServer) AddPlayer(_ context.Context, req *gen.AddPlayerRequest) (*gen.AddPlayerResponse, error) {
	if len(req.Players) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "no players provided")
	}

	players := make([]matchmaking.Player, 0, len(req.Players))
	for _, p := range req.Players {
		players = append(players, matchmaking.Player{
			ID:    p.Id,
			Level: int(p.Level),
		})
	}

	s.service.AddPlayer(players...)

	return &gen.AddPlayerResponse{}, nil
}

func (s *MatchmakingServer) RemovePlayer(_ context.Context, req *gen.RemovePlayerRequest) (*gen.RemovePlayerResponse, error) {
	if len(req.Players) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "no players provided")
	}

	players := make([]matchmaking.Player, 0, len(req.Players))
	for _, p := range req.Players {
		players = append(players, matchmaking.Player{
			ID:    p.Id,
			Level: int(p.Level),
		})
	}

	s.service.RemovePlayer(players...)

	return &gen.RemovePlayerResponse{}, nil
}

func (s *MatchmakingServer) Status(req *gen.StatusRequest, stream grpc.ServerStreamingServer[gen.StatusResponse]) error {
	s.logger.Debug("Status request", slog.Any("request", req))

	// TODO: check if player exists and authenticated
	s.l.Lock()
	s.playerStates[req.PlayerId] = stream
	s.l.Unlock()
	metrics.OnlinePlayers.Inc()
	defer func() {
		s.l.Lock()
		delete(s.playerStates, req.PlayerId)
		s.l.Unlock()
		metrics.OfflinePlayers.Inc()
	}()

	<-stream.Context().Done()

	return nil
}

// RunStatusUpdater sends status updates to players
func (s *MatchmakingServer) RunStatusUpdater(ctx context.Context, outputStatus <-chan matchmaking.MatchSession) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case match := <-outputStatus:
			metrics.TotalPlayers.WithLabelValues(match.Type).Add(float64(len(match.Players)))
			s.logger.DebugContext(ctx, "Player status updater:", slog.String("type", match.Type), slog.Any("players", match.Players))

			for _, player := range match.Players {
				s.l.RLock()
				stream, ok := s.playerStates[player.ID]
				s.l.RUnlock()
				if ok {
					resp := &gen.StatusResponse{
						Id:      match.ID,
						Created: timestamppb.New(match.Created),
						Type:    match.Type,
					}
					if match.Type == matchmaking.ChangesTypeMatchFound {
						resp.Players = make([]*gen.PlayerData, 0, len(match.Players))
						for _, p := range match.Players {
							resp.Players = append(resp.Players, &gen.PlayerData{
								Id:    p.ID,
								Level: int32(p.Level),
							})
						}
					}

					err := stream.Send(resp)
					if err != nil {
						s.logger.DebugContext(ctx, "failed to send status", slog.String("player_id", player.ID), slog.String("error", err.Error()))
					}
				}
			}
		}
	}

	return nil
}
