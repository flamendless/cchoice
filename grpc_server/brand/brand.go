package brand

import (
	"cchoice/internal/ctx"
	"cchoice/internal/serialize"
	pb "cchoice/proto"
	"context"
)

type BrandServer struct {
	pb.UnimplementedBrandServiceServer
	CtxDB *ctx.Database
}

func NewGRPCBrandServer(ctxDB *ctx.Database) *BrandServer {
	return &BrandServer{CtxDB: ctxDB}
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
