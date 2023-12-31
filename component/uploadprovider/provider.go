package uploadprovider

import (
	"bestHabit/common"
	"context"
)

type UploadProvider interface {
	SaveFileUploaded(ctx context.Context, data []byte, dst string) (*common.Image, error)
}
