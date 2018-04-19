package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func main() {
	fmt.Println("Main Called")
	diff()
}

func diff() {
	lf, err := os.Open("lhs.txt")
	rf, err := os.Open("rhs.txt")
	defer lf.Close()
	defer rf.Close()
	if err != nil {
		log.Fatalf("unable to open file: %v", err)
	}
	lhsLines := parse(lf)
	rhsLines := parse(rf)
	// strings.Fields()
	lart := make(chan Artifact, 10)
	rart := make(chan Artifact, 10)
	go constructObj(lhsLines, lart)
	go constructObj(rhsLines, rart)
	var larts = make(map[string]string)
	var rarts = make(map[string]string)
	ldone := make(chan bool)
	rdone := make(chan bool)
	go func(done chan bool) {
		for artifact := range lart {
			larts[artifact.component] = artifact.version
			// larts = append(larts, artifact)
		}
		log.Println("ldone")
		ldone <- true
	}(ldone)
	go func(done chan bool) {
		for artifact := range rart {
			rarts[artifact.component] = artifact.version
			// rarts = append(rarts, artifact)
		}
		log.Println("rdone")
		rdone <- true
	}(rdone)
	var isLeftDone, isRighDone = false, false
	for {
		select {
		case <-ldone:
			isLeftDone = true
		case <-rdone:
			isRighDone = true
		}
		if isLeftDone && isRighDone {
			break
		}
	}
	fullArtifacts := compare(larts, rarts)
	constructTable(fullArtifacts)
}

func constructTable(artifacts FullArtifacts) {
	f, err := os.Create("out.txt")
	if err != nil {
		log.Fatalf("unable to create file: %v", err)
	}
	table := tablewriter.NewWriter(f)
	table.SetHeader([]string{"Component", "Left Version", "Right Version", "Is Different"})
	for _, data := range artifacts {
		var d []string
		if data.LeftVersion != data.RightVersion {
			d = []string{data.Component, data.LeftVersion, data.RightVersion, "Yes"}
		} else {
			d = []string{data.Component, data.LeftVersion, data.RightVersion, "No"}
		}
		table.Append(d)
	}
	table.Render()
}

func compare(lhs, rhs map[string]string) FullArtifacts {
	var fullArtifacts FullArtifacts
	for key, val := range lhs {
		fullArtifact := FullArtifact{}
		if rval, ok := rhs[key]; ok {
			fullArtifact.Component = key
			fullArtifact.LeftVersion = val
			fullArtifact.RightVersion = rval
			if fullArtifact.LeftVersion != fullArtifact.RightVersion {
				fullArtifact.IsDiff = true
			} else {
				fullArtifact.IsDiff = false
			}
		}
		fullArtifacts = append(fullArtifacts, fullArtifact)
	}
	return fullArtifacts
}

type FullArtifact struct {
	Component    string
	LeftVersion  string
	RightVersion string
	IsDiff       bool
}

type FullArtifacts []FullArtifact

type Artifact struct {
	component string
	version   string
}

type Artifacts []Artifact

func parse(f *os.File) []string {
	scanner := bufio.NewScanner(f)
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}
	return lines
}

func constructObj(lines []string, art chan Artifact) {
	for _, line := range lines {
		artifact := Artifact{}
		tokens := strings.Fields(line)
		for i, token := range tokens {
			if i == 0 {
				artifact.component = token
			}
			if i == 1 {
				artifact.version = token
			}
		}
		art <- artifact
	}
}
