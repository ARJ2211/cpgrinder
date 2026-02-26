package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/ARJ2211/cpgrinder/internal/platform"
	"github.com/ARJ2211/cpgrinder/internal/store"
)

type globalFlags struct {
	DB     string
	Import string
	Reset  bool
}

func PrintJSON(obj interface{}) {
	bytes, _ := json.MarshalIndent(obj, "\t", "\t")
	fmt.Println(string(bytes))
}

func main() {
	var gf globalFlags

	flag.StringVar(&gf.DB, "db", "", "use custom DB path to sqlite")
	flag.StringVar(&gf.Import, "import", "", "force re-import of db")
	flag.BoolVar(&gf.Reset, "reset", false, "deletes the database for a fresh start")

	flag.Parse()

	if gf.DB != "" {
		fmt.Println("using new db path: " + gf.DB)
	}

	if gf.Import != "" {
		fmt.Println("force re-importing from: " + gf.Import)
	}

	var confBreak bool = !gf.Reset
	for !confBreak {
		var confirmation string
		fmt.Print("are you sure you want to reset your progress (y/n): ")
		fmt.Scanln(&confirmation)
		fmt.Println()

		switch confirmation {
		case "y", "Y", "Yes":
			fmt.Println("resetting progress...")
			confBreak = true
			// TODO: Reset db here...
		case "n", "N", "No":
			fmt.Println("continuing...")
			confBreak = true
		default:
			fmt.Println("incorrect option provided")
		}
	}

	res := fmt.Sprintf(`
	You have selected the following flags:
	--------------------------------------
	1. DB path: %s
	2. Import path: %s
	3. Reset bool: %v
	`, gf.DB, gf.Import, gf.Reset)

	println(res + "\n\n")

	dbPath, workspaceDir, err := platform.ResolvePaths(gf.DB)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println(dbPath)
	fmt.Println(workspaceDir)

	dbStore, err := store.Open(dbPath)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Count of the problems at fresh db
	c, err := dbStore.CountProblems()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println("COUNT OF PROBLEMS: " + strconv.Itoa(c))

	// Upsert the fixtures from the catalog.json
	if err := dbStore.UpsertProblemsFromFixture(gf.Import); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	c1, err := dbStore.CountProblems()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println("COUNT OF UPSERTED PROBLEMS: " + strconv.Itoa(c1))

	// List Problems that are there in the database
	fmt.Println()
	uf := store.UserFilters{Limit: 10, Title: "elevator"}
	problems, err := dbStore.ListProblems(uf)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	for i, p := range problems {
		fmt.Println(strconv.Itoa(i + 1))
		fmt.Println(p.Title, p.Id)
	}

	// Get a problem by id
	fmt.Println()
	pID, err := dbStore.GetProblemByID("7c7b07f7-0e6f-40d8-a55a-0a160ac1ef56")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	PrintJSON(pID)

	// CLOSE THE DB
	if err := dbStore.Close(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
