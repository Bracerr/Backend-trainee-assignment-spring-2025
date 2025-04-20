package grpc

import (
	"context"
	"time"

	pb "avito-backend/src/internal/delivery/grpc/pb"
	"avito-backend/src/internal/service"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type PVZGrpcServer struct {
	pb.UnimplementedPVZServiceServer
	pvzService service.PVZServiceInterface
}

func NewPVZGrpcServer(pvzService service.PVZServiceInterface) *PVZGrpcServer {
	return &PVZGrpcServer{
		pvzService: pvzService,
	}
}

func (s *PVZGrpcServer) GetPVZList(ctx context.Context, req *pb.GetPVZListRequest) (*pb.GetPVZListResponse, error) {
	pvzList, err := s.pvzService.GetPVZsWithReceptions(time.Time{}, time.Time{}, 0, 1000)
	if err != nil {
		return nil, err
	}

	response := &pb.GetPVZListResponse{
		Pvzs: make([]*pb.PVZ, 0, len(pvzList)),
	}

	for _, pvz := range pvzList {
		response.Pvzs = append(response.Pvzs, &pb.PVZ{
			Id:               pvz.PVZ.ID.String(),
			RegistrationDate: timestamppb.New(pvz.PVZ.RegistrationDate),
			City:             string(pvz.PVZ.City),
		})
	}

	return response, nil
}
