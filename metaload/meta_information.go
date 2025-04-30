/*
 *@author  chengkenli
 *@project setbuckets
 *@package metaload
 *@file    meta_information
 *@date    2025/4/17 9:55
 */

package metaload

/*
CREATE TABLE IF NOT EXISTS ops.starrocks_devops_optimize_information (
`ts`           date               COMMENT '分区',
`app`          varchar(200)       COMMENT '集群名称',
`database`     varchar(200)       COMMENT '库名',
`table`        varchar(500)       COMMENT '表名',
`stmt_before`  varchar(1048576)   COMMENT '调整前-语句',
`stmt_after`   varchar(1048576)   COMMENT '调整后-语句',
`operational`  varchar(500)       COMMENT '行为',
`edtime`       varchar(100)       COMMENT '耗时',
`state`        varchar(500)       COMMENT '状态',
`timestamp`    datetime           COMMENT '操作时间',
`comment`      varchar(500)       COMMENT '备注'
) ENGINE=OLAP
DUPLICATE KEY(`ts`, `app`)
COMMENT "管理员副本治理明细表"
PARTITION BY DATE_TRUNC('DAY',ts)
DISTRIBUTED BY HASH(`ts`,`app`) BUCKETS 5
*/

type MetaInfo struct {
	Ts          string `json:"ts"`
	App         string `json:"app"`
	Database    string `json:"database"`
	Table       string `json:"table"`
	StmtBefore  string `json:"stmt_before"`
	StmtAfter   string `json:"stmt_after"`
	Operational string `json:"operational"`
	Edtime      string `json:"edtime"`
	State       string `json:"state"`
	Timestamp   string `json:"timestamp"`
	Comment     string `json:"comment"`
}
