package fa_score_analysis

import (
	"context"
	"dataset/db"
	"dataset/fetch"
	"fmt"
	"testing"
)

func TestGetFAScoreThresholds(t *testing.T) {
	ctx := context.Background()
	user, _ := fetch.GetTestUser()
	list := []string{"N2KTB_ESB", "N2CFM_BSM", "N2CHF_TBL", "N2CUL_MNT", "N2ENGWEB"}
	for _, database := range list {
		conn, status := db.NewerDBAdapter(ctx, false, user.Username, database)
		if status.IsErr {
			t.Fatal(status)
		}
		critical, question, status := GetFAScoreThresholds(conn)
		if status.IsErr {
			t.Fatal(status)
		}
		fmt.Println(database, "Critical:", critical, "Question:", question)
		conn.Close()
	}
}

func TestFAScoreAnalysis(t *testing.T) {
	ctx := context.Background()
	user, _ := fetch.GetTestUser()
	list := []string{"N2KTB_ESB", "N2CFM_BSM", "N2CHF_TBL", "N2CUL_MNT"}
	for _, database := range list {
		conn, status := db.NewerDBAdapter(ctx, false, user.Username, database)
		if status.IsErr {
			t.Fatal(status)
		}
		output, status := FAScoreAnalysis(conn)
		if status.IsErr {
			t.Fatal(status)
		}
		conn.Close()
		fmt.Println(output)
	}
}
