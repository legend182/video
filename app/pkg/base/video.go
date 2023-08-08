package base

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/qinguoyi/osproxy/app/pkg/repo"
	"github.com/qinguoyi/osproxy/app/pkg/utils"
	"github.com/qinguoyi/osproxy/bootstrap/plugins"
)

func InitVideo() {
	lgDB := new(plugins.LangGoDB).Use("default").NewDB()
	video, err := repo.NewMetaDataInfoRepo().GetAllVideo(lgDB)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	redis := new(plugins.LangGoRedis).NewRedis()
	exists := redis.Exists(context.Background(), utils.Video).Val()
	if exists == 1 {
		result := redis.Del(context.Background(), utils.Video).Val()
		if result == 0 {
			panic("删除video失败")
		}
	}
	for _, info := range video {
		marshal, err := json.Marshal(info)
		if err != nil {
			panic(err)
		}
		push := redis.RPush(context.Background(), utils.Video, marshal)
		if push.Err() != nil {
			fmt.Println(push.Err())
			panic(err)
		}
	}
}
