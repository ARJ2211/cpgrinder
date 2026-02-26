package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/ARJ2211/cpgrinder/internal/platform"
	"github.com/ARJ2211/cpgrinder/internal/store"
	"github.com/ARJ2211/cpgrinder/tui/problem_list"
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

	dbPath, workspacePath, err := platform.ResolvePaths(gf.DB)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	dbStore, err := store.Open(dbPath, workspacePath)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	m, err := problem_list.InitialModel(dbStore)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Clear terminal screen
	fmt.Print("\033[H\033[2J")

	// uf := store.UserFilters{
	// 	Limit:  10,
	// 	Offset: 20,
	// }
	// p, err := dbStore.ListProblems(uf)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	os.Exit(1)
	// }
	// for _, s := range p {
	// 	fmt.Println(s.Id)
	// }

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
