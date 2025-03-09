package service

import (
	"bytes"
	"context"
	"io"
	"log"
	"pcbook/pb"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxImageSize = 1 << 20

type LaptopServer struct {
	laptopStore LaptopStore
	imageStore  ImageStore
	ratingStore RatingStore
	pb.UnimplementedLaptopServiceServer
}

func NewLaptopServer(laptopStore LaptopStore, imageStore ImageStore, ratingStore RatingStore) *LaptopServer {
	return &LaptopServer{laptopStore: laptopStore, imageStore: imageStore, ratingStore: ratingStore}
}

func (server *LaptopServer) CreateLaptop(
	ctx context.Context,
	req *pb.CreateLaptopRequest,
) (*pb.CreateLaptopResponse, error) {
	laptop := req.GetLaptop()
	log.Printf("receive a create-laptop request with id: %s", laptop.Id)

	if len(laptop.Id) > 0 {
		_, err := uuid.Parse(laptop.Id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "laptop ID is not a valid UUID: %v", err)
		}
	} else {
		id, err := uuid.NewRandom()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "cannot generate a new laptop ID: %v", err)
		}
		laptop.Id = id.String()
	}

	// save the laptop to the store
	err := server.laptopStore.Save(laptop)
	if err != nil {
		code := codes.Internal
		if err == ErrAlreadyExists {
			code = codes.AlreadyExists
		}
		return nil, status.Errorf(code, "cannot save laptop to the store: %v", err)
	}

	res := &pb.CreateLaptopResponse{
		Id: laptop.Id,
	}
	log.Printf("laptop with id %s has been saved", laptop.Id)
	return res, nil
}

func (server *LaptopServer) SearchLaptop(
	req *pb.SearchLaptopRequest,
	stream pb.LaptopService_SearchLaptopServer,
) error {
	filter := req.GetFilter()
	log.Printf("receive a search-laptop request with filter: %v", filter)

	err := server.laptopStore.Search(
		stream.Context(),
		filter,
		func(laptop *pb.Laptop) error {
			res := &pb.SearchLaptopResponse{Laptop: laptop}
			err := stream.Send(res)
			if err != nil {
				return err
			}

			log.Printf("sent laptop with id: %s", laptop.GetId())
			return nil
		},
	)

	if err != nil {
		return status.Errorf(codes.Internal, "unexpected error: %v", err)
	}

	return nil
}

func (server *LaptopServer) UploadImage(stream pb.LaptopService_UploadImageServer) error {
	req, err := stream.Recv()
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot receive image info: %v", err))
	}

	laptopID := req.GetInfo().GetLaptopId()
	imageType := req.GetInfo().GetImageType()
	log.Printf("receive an upload-image request for laptop %s with type %s", laptopID, imageType)

	laptop, err := server.laptopStore.Find(laptopID)
	if err != nil {
		return logError(status.Errorf(codes.NotFound, "cannot find laptop: %v", err))
	}
	if laptop == nil {
		return logError(status.Errorf(codes.NotFound, "laptop %s is not found", laptopID))
	}

	imageData := bytes.Buffer{}
	imageSize := 0

	for {
		err := contextError(stream.Context())
		if err != nil {
			return err
		}

		log.Printf("waitting for image chunk...")
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return logError(status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err))
		}

		chunk := req.GetChunkData()
		imageSize += len(chunk)

		if imageSize > maxImageSize {
			return logError(
				status.Errorf(
					codes.InvalidArgument,
					"image is too large: %d > %d",
					imageSize,
					maxImageSize,
				),
			)
		}
		_, err = imageData.Write(chunk)
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "cannot write chunk data: %v", err))
		}
	}

	imageID, err := server.imageStore.Save(laptopID, imageType, imageData)
	if err != nil {
		return logError(status.Errorf(codes.Internal, "cannot save image to the store: %v", err))
	}

	res := &pb.UploadImageResponse{
		Id:   imageID,
		Size: uint32(imageSize),
	}
	err = stream.SendAndClose(res)
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot send response: %v", err))
	}
	log.Printf("saved image with id: %s", imageID)
	return nil
}

func (server *LaptopServer) RateLaptop(stream pb.LaptopService_RateLaptopServer) error {
	for {
		error := contextError(stream.Context())
		if error != nil {
			return error
		}

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data in request")
			break
		}

		if err != nil {
			logError(status.Errorf(codes.Unknown, "cannot receive request: %v", err))
			break
		}

		laptopID := req.GetLaptopId()
		score := req.GetScore()
		log.Printf("receive a rate-laptop request for laptop %s with score %f", laptopID, score)

		found, err := server.laptopStore.Find(laptopID)
		if err != nil {
			logError(status.Errorf(codes.Internal, "cannot find laptop: %v", err))
			break
		}
		if found == nil {
			logError(status.Errorf(codes.NotFound, "laptop %s is not found", laptopID))
			break
		}

		rating, err := server.ratingStore.Add(laptopID, score)
		if err != nil {
			logError(status.Errorf(codes.Internal, "cannot add rating to the store: %v", err))
			break
		}

		res := &pb.RateLaptopResponse{
			LaptopId:     laptopID,
			RatedCount:   rating.Count,
			AverageScore: rating.Sum / float64(rating.Count),
		}
		err = stream.Send(res)
		if err != nil {
			logError(status.Errorf(codes.Unknown, "cannot send response: %v", err))
			break
		}
	}
	return nil
}

func logError(err error) error {
	if err != nil {
		log.Print(err)
	}
	return err
}

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return logError(status.Error(codes.Canceled, "request is canceled"))
	case context.DeadlineExceeded:
		return logError(status.Error(codes.DeadlineExceeded, "deadline is exceeded"))
	default:
		return nil
	}
}
