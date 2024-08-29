package brand

import (
	"cchoice/internal/ctx"
	"cchoice/internal/serialize"
	pb "cchoice/proto"
	"context"
	"errors"
	"os"
)

type BrandServer struct {
	pb.UnimplementedBrandServiceServer
	CtxDB *ctx.Database
}

func NewGRPCBrandServer(ctxDB *ctx.Database) *BrandServer {
	return &BrandServer{CtxDB: ctxDB}
}

func (s *BrandServer) GetBrand(
	ctx context.Context,
	in *pb.GetBrandRequest,
) (*pb.GetBrandResponse, error) {
	brandID := serialize.DecDBID(in.Id)
	brand, err := s.CtxDB.Queries.GetBrandByID(context.Background(), brandID)
	if err != nil {
		return nil, err
	}
	return &pb.GetBrandResponse{
		Brand: &pb.Brand{
			Id: in.Id,
			Name: brand.Name,
			MainImage: &pb.BrandImage{
				Id:      serialize.EncDBID(brand.BrandImageID),
				BrandId: in.Id,
				Path:    brand.Path,
			},
		},
	}, nil
}

func (s *BrandServer) GetBrandLogos(
	ctx context.Context,
	in *pb.GetBrandLogosRequest,
) (*pb.GetBrandLogosResponse, error) {
	brandLogos, err := s.CtxDB.Queries.GetBrandLogos(context.Background(), 100)
	if err != nil {
		return nil, err
	}

	brands := make([]*pb.Brand, 0, len(brandLogos))
	for _, brandLogo := range brandLogos {
		if len(brands) >= int(in.Limit) {
			break
		}

		_, err := os.Stat("client/" + brandLogo.Path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}

		serBrandID := serialize.EncDBID(brandLogo.ID)
		brands = append(brands, &pb.Brand{
			Id:   serBrandID,
			Name: brandLogo.Name,
			MainImage: &pb.BrandImage{
				Id:      serialize.EncDBID(brandLogo.BrandImageID),
				BrandId: serBrandID,
				Path:    brandLogo.Path,
				IsMain:  true,
			},
		})
	}

	return &pb.GetBrandLogosResponse{
		Length: int64(len(brandLogos)),
		Brands: brands,
	}, nil
}
