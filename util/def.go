package util

import (
	"github.com/sirupsen/logrus"
	"time"
)

var (
	P      Parms
	Loggrs *logrus.Logger
	// 全局StarRocks连接配置
	SrCfg SrAvgs
)

type Parms struct {
	Help           bool
	App            string
	Action         string
	Buckets        int64
	List           bool
	BackendId      string
	ClearPartition bool
	Thread         int
	Auto           bool
	SortKey        string
	SplitKey       string
	Table          string
	TabletSize     int64
	PartitionSet   bool
	PartitionName  string
}

type SrAvgs struct {
	Host string
	Port int
	User string
	Pass string
}

type Fix struct {
	App      string `bson:"app"`
	Database string `bson:"database"`
	Table    string `bson:"table"`
	Count    int64  `bson:"count"`
	Edtime   int64  `bson:"edtime"`
	Before   int64  `bson:"before"`
	Last     int64  `bson:"last"`
	Comment  string `bson:"comment"`
}

type Catch []struct {
	Ts     time.Time `bson:"ts"`
	App    string    `bson:"app"`
	User   string    `bson:"user"`
	Grants string    `bson:"grants"`
}
