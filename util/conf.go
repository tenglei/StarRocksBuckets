package util

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"os"
	"path/filepath"
	"strings"
)

func usage() {
	fmt.Printf("\nUsage: %s [-s adhoc] [-h]\n\nOptions:\nStarRocks Buckets分桶修复工具\n", filepath.Base(os.Args[0]))
	flag.PrintDefaults()
	fmt.Println()
}

func init() {
	c := color.New()
	flag.StringVar(&P.Action, "m", "create,insert,alter,drop",
		fmt.Sprintf("%s,%s,%s,%s",
			c.Add(color.FgHiGreen).Sprint("create"),
			c.Add(color.FgHiWhite).Sprint("insert"),
			c.Add(color.FgHiYellow).Sprint("alter"),
			c.Add(color.FgHiYellow).Sprint("drop"),
		))
	flag.StringVar(&P.Table, "t", "", fmt.Sprintf("<%s>", c.Add(color.FgHiWhite).Sprint("SCHEMA.TABLE")))
	flag.Int64Var(&P.Buckets, "b", 0, c.Add(color.FgHiWhite).Sprint("BUCKETS"))
	flag.StringVar(&P.BackendId, "id", "", c.Add(color.FgHiWhite).Sprint("BACKEND ID"))
	flag.StringVar(&P.SplitKey, "splitkey", "", c.Add(color.FgHiCyan).Sprint("BUCKET KEY"))
	flag.StringVar(&P.SortKey, "sortkey", "", c.Add(color.FgHiCyan).Sprint("SORT KEY"))
	flag.StringVar(&P.App, "s", "", "<APP>")
	flag.BoolVar(&P.Help, "h", false, "show help information")
	flag.BoolVar(&P.List, "l", false, c.Add(color.FgHiWhite).Sprint("SHOW SORT KEY"))
	flag.BoolVar(&P.Auto, "a", false, "AUTO OVERWRITE")
	flag.BoolVar(&P.ClearPartition, "c", false, c.Add(color.FgHiRed).Sprint("FILTER EMPTY PARTITIONS"))
	flag.Int64Var(&P.TabletSize, "n", 0, fmt.Sprintf("%s/MB", c.Add(color.FgHiWhite).Sprint("SPLIT TABLET")))
	flag.IntVar(&P.Thread, "thread", 3, c.Add(color.FgHiYellow).Sprint("THREAD"))
	flag.BoolVar(&P.PartitionSet, "pset", false, fmt.Sprintf("[%s] %s LEVEL RESET", c.Add(color.FgHiYellow).Sprint("PARTITION"), c.Add(color.FgHiWhite).Sprint("BUCKET")))
	flag.StringVar(&P.PartitionName, "p", "", fmt.Sprintf("%s Name", c.Add(color.FgHiYellow).Sprint("PARTITIONS")))

	flag.Parse()
	flag.Usage = usage

	if !strings.Contains(P.Table, ".") {
		flag.Usage()
		os.Exit(1)
	}
	if P.Help || len(P.Table) == 0 {
		flag.Usage()
		os.Exit(-1)
	}
	Logrus()

}

func Parm() {
}
