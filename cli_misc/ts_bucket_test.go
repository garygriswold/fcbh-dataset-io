package cli_misc

import (
	"context"
	"fmt"
	"testing"
)

func TestTSBucket(t *testing.T) {
	ctx := context.Background()
	ts := NewTSBucket(ctx)
	key := ts.GetKey(ScriptTS, `ENGWEBN2DA`, `REV`, 22)
	fmt.Println(key)
	object := ts.GetObject(TSBucketName, key)
	fmt.Println(string(object))
	timestamps := ts.GetTimestamps(ScriptTS, `ENGWEBN2DA`, `REV`, 22)
	for _, time := range timestamps {
		fmt.Println(time)
	}
}
