package usergrpcclient

import (
	"bestHabit/common"
	proto "bestHabit/generatedProto/proto/userservice"
	"bestHabit/modules/user/usermodel"
	"context"
	"fmt"
)

type gRPCClient struct {
	client proto.UserServiceClient
}

func NewGRPCClient(client proto.UserServiceClient) *gRPCClient {
	return &gRPCClient{client: client}
}

func (c *gRPCClient) UpdateUserInfoByGRPC(ctx context.Context, userId int, userUpdate *usermodel.UserUpdate) (*usermodel.User, error) {
	res, err := c.client.UserUpdateProfile(ctx, &proto.UserUpdateProfileRequest{
		UserId: int32(userId),
		Phone:  *userUpdate.Phone,
		Name:   *userUpdate.Name,
		Avatar: &proto.Image{
			Id:        int32(userUpdate.Avatar.Id),
			Url:       userUpdate.Avatar.Url,
			CloudName: userUpdate.Avatar.CloudName,
			Extension: userUpdate.Avatar.Extension,
			Width:     int32(userUpdate.Avatar.Width),
			Height:    int32(userUpdate.Avatar.Height),
		},
		Settings: &proto.Settings{
			Theme:    userUpdate.Settings.Theme,
			Language: userUpdate.Settings.Language,
		},
	})

	fmt.Println("CC 1")

	if err != nil {
		return nil, common.ErrDB(err)
	}

	fmt.Println("CC 2")

	return &usermodel.User{
		SQLModel: common.SQLModel{
			Id: int(res.UserId),
		},
		Name:  &res.Name,
		Email: &res.Email,
		FbID:  &res.FbId,
		GgID:  &res.GgId,
		Avatar: &common.Image{
			Url:       res.Avatar.Url,
			CloudName: res.Avatar.CloudName,
			Extension: res.Avatar.Extension,
			Height:    int(res.Avatar.Height),
			Width:     int(res.Avatar.Width),
		},
		Settings: &common.Settings{
			Theme:    res.Settings.Theme,
			Language: res.Settings.GetLanguage(),
		},
	}, nil
}
