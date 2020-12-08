package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func parseCorpus(fileReader io.Reader, charsMap map[string]int, digraphsMap map[string]int) error {

	reader := bufio.NewReader(fileReader)

	prev_c := ""
	prev_c_dk := ""

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			break
		}

		for _, c := range line {
			if int(c) > 31 {

				s := mapToAscii(strings.ToLower(string(c)))

				if strings.Contains("äëïüöâûîô", s) {
					dk := getDeadKey(s)
					addDeadKeyChar(dk, string(prev_c), charsMap, digraphsMap)
					prev_c_dk = dk.livechar
				} else {
					addToMapAndIncrement(s, charsMap)

					if prev_c != "" {
						if prev_c_dk != "" {
							prev_c = prev_c_dk
						}
						addToMapAndIncrement(fmt.Sprintf("%s%s", prev_c, s), digraphsMap)
					}
					prev_c_dk = ""
				}

				prev_c = s

			}
		}

		if err != nil {
			break
		}
	}
	return nil
}

func mapToAscii(s string) string {
	if strings.Contains("()[]{}", s) {
		return " "
	}
	switch s {
	case "é":
		return ")"
	case "è":
		return "("
	case "à":
		return "{"
	case "ê":
		return "}"
	case "ç":
		return "["
	case "ù":
		return "]"
	default:
		return s
	}
}

func addToMapAndIncrement(s string, m map[string]int) {
	if i, found := m[s]; found {
		m[s] = i + 1
	} else {
		m[s] = 1
	}
}

func addDeadKeyChar(dk deadKey, prev string, charsMap map[string]int, digraphsMap map[string]int) {
	addToMapAndIncrement(dk.deadchar, charsMap)
	addToMapAndIncrement(dk.livechar, charsMap)
	addToMapAndIncrement(fmt.Sprintf("%s%s", prev, dk.deadchar), digraphsMap)
	addToMapAndIncrement(fmt.Sprintf("%s%s", dk.deadchar, dk.livechar), digraphsMap)
}

type deadKey struct {
	deadchar string
	livechar string
}

func getDeadKey(c string) deadKey {
	switch c {
	case "ä":
		return deadKey{"¨", "a"}
	case "ï":
		return deadKey{"¨", "i"}
	case "ë":
		return deadKey{"¨", "e"}
	case "ü":
		return deadKey{"¨", "u"}
	case "ö":
		return deadKey{"¨", "o"}
	case "â":
		return deadKey{"^", "a"}
	case "î":
		return deadKey{"^", "a"}
	case "ê":
	//  ê so frequent is a key itself !
	// return deadKey{"^", "e"}
	case "û":
		return deadKey{"^", "u"}
	case "ô":
		return deadKey{"^", "o"}
	}
	return deadKey{}
}

func FilePathWalkDir(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func main() {
	flag.Parse()
	if len(flag.Args()) == 0 {
		fmt.Printf("usage: cByC <file1> [<file2> ...]\n")
		return
	}

	charsFileWriter, err := os.Create("chars.txt")
	if err != nil {
		fmt.Print(err)
	}
	defer charsFileWriter.Close()

	digraphsFileWriter, err := os.Create("digraphs.txt")
	if err != nil {
		fmt.Print(err)
	}
	defer digraphsFileWriter.Close()

	files, err := FilePathWalkDir(flag.Args()[0])
	if err != nil {
		panic(err)
	}

	charsWriter := bufio.NewWriter(charsFileWriter)
	digraphsWriter := bufio.NewWriter(digraphsFileWriter)

	charsMap := make(map[string]int)
	digraphsMap := make(map[string]int)

	for _, file := range files {

		fileReader, err := os.Open(file)
		if err != nil {
			fmt.Print(err)
		}
		defer fileReader.Close()

		err = parseCorpus(fileReader, charsMap, digraphsMap)
		if err != nil {
			fmt.Println(err)
		}
	}

	sortedChars := getSortedMap(charsMap)

	for _, kv := range sortedChars {
		charsWriter.WriteString(fmt.Sprintf("%s %d\n", kv.Key, kv.Value))
	}

	sortedDigraphs := getSortedMap(digraphsMap)
	for _, kv := range sortedDigraphs {
		digraphsWriter.WriteString(fmt.Sprintf("%s %d\n", kv.Key, kv.Value))
	}

	charsWriter.Flush()
	digraphsWriter.Flush()

}

type entry struct {
	Key   string
	Value int
}

func getSortedMap(m map[string]int) []entry {
	var entries []entry
	for k, v := range m {
		entries = append(entries, entry{k, v})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Value > entries[j].Value
	})
	return entries
}
