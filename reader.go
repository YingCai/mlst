package mlst

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// extend the bufio.Reader, and keep track of the line number
type Reader struct {
	reader  *bufio.Reader
	lineNum int
	caseNum int
	line    string
}
type Error struct {
	lineNum int
	caseNum int
	line    string
	message string
}

type InFileReader struct {
	Reader
}
type OutFileReader struct {
	Reader
}
type TeamInfoReader struct {
	Reader
}

/*** Constants and Initializations ***/

const (
	InvalidNumLeaves = -1
)

var (
	NumberFormatRegexp [3]*regexp.Regexp
	NumberExpectedMsg  [3]string
	LoginFormatRegexp  *regexp.Regexp
	LoginExpectedMsg   string
)

func init() {
	NumberFormatRegexp[1] = regexp.MustCompile("^(\\d+)\\n?$")
	NumberFormatRegexp[2] = regexp.MustCompile("^(\\d+) (\\d+)\\n?$")
	NumberExpectedMsg[1] = "Expecting one natural number on a line with " +
		"no leading or trailing spaces."
	NumberExpectedMsg[2] = "Expecting two natural numbers separated by a " +
		"single space on a line with no leading or trailing spaces."

	LoginFormatRegexp = regexp.MustCompile("^[[:alpha:]]{2}$")
	LoginExpectedMsg = "Expecting two letters for login."
}

/*** Error ***/

func (e *Error) caseInfo() string {
	if e.caseNum == 0 {
		return ""
	}
	return fmt.Sprintf(" (Graph #%d)", e.caseNum)
}

func (e *Error) Error() string {
	return fmt.Sprintf("Error on line %d%s: Got '%s'. %s", e.lineNum,
		e.caseInfo(), e.line, e.message)
}

/*** Reader ***/

func (r *Reader) readLine() (err error) {
	r.lineNum++
	r.line, err = r.reader.ReadString('\n')
	if len(r.line) > 0 {
		err = nil
	}
	return
}

func (r *Reader) strippedLine() string {
	if len(r.line) > 0 && r.line[len(r.line)-1] == '\n' {
		return r.line[:len(r.line)-1]
	}
	return r.line
}

// construct a new Error struct
func (r *Reader) Error(msg string) *Error {
	return &Error{lineNum: r.lineNum, caseNum: r.caseNum, line: r.strippedLine(),
		message: msg}
}

// construct a new Error struct with the expected message
func (r *Reader) ErrorWithExpected(msg, expected string) *Error {
	return r.Error(msg + " " + expected)
}

// read one or two numbers on a line
func (r *Reader) readNumbers(msg string, n int) ([]int, *Error) {
	if err := r.readLine(); err != nil {
		return nil, r.ErrorWithExpected(msg+" ("+err.Error()+")", "Expecting the next line.")
	}

	matches := NumberFormatRegexp[n].FindStringSubmatch(r.line)
	if matches == nil || len(matches) != n+1 {
		return nil, r.ErrorWithExpected(msg, NumberExpectedMsg[n])
	}

	ints := make([]int, len(matches)-1)
	for i := range ints {
		if num, err := strconv.Atoi(matches[i+1]); err != nil {
			return nil, r.ErrorWithExpected(msg+" ("+err.Error()+")", NumberExpectedMsg[n])
		} else {
			ints[i] = num
		}
	}
	return ints, nil
}

/*** InFileReader ***/

func NewInFileReader(rd io.Reader) *InFileReader {
	var inReader InFileReader
	inReader.reader = bufio.NewReader(rd)
	return &inReader
}

func (r *InFileReader) ReadInputFile() ([]EdgeSet, *Error) {
	nums, err := r.readNumbers("Cannot parse the number of input graphs.", 1)
	if err != nil {
		return nil, err
	}

	NumCases := nums[0]
	edgeSets := make([]EdgeSet, NumCases)
	for r.caseNum = 1; r.caseNum <= NumCases; r.caseNum++ {
		edgeSet, err := r.ReadInputGraph()
		if err != nil {
			return nil, err
		}

		G := edgeSet.Graph()
		G.Search()
		if !G.EdgesInOneComponent() {
			return nil, r.Error("Disconnected graph: after reading the last edge of " +
				"this graph, the edges are not in the same component.")
		}

		edgeSets[r.caseNum-1] = edgeSet
	}

	if err := r.readLine(); err != io.EOF {
		return nil, r.ErrorWithExpected(fmt.Sprintf(
			"Extra lines after Graph #%d (line 1 says the number of graphs is %d).",
			NumCases, NumCases), "Expecting EOF.")
	}
	return edgeSets, nil
}

func (r *InFileReader) ReadInputGraph() (EdgeSet, *Error) {
	nums, err := r.readNumbers("Cannot parse the number of edges.", 1)
	if err != nil {
		return nil, err
	}

	NumEdges := nums[0]
	if NumEdges > MaxNumEdges {
		return nil, r.Error(fmt.Sprintf("Number of edges cannot exceed %d.", MaxNumEdges))
	}

	edgeSet := make(EdgeSet)
	for i := 0; i < NumEdges; i++ {
		nums, err = r.readNumbers("Cannot parse the next edge.", 2)
		if err != nil {
			return nil, err
		} else {
			var e Edge
			e.Ends[0] = nums[0]
			e.Ends[1] = nums[1]
			e.Normalize()

			if err := e.Error(); err != nil {
				return nil, r.Error(err.Error())
			}

			if _, ok := edgeSet[e]; ok {
				return nil, r.Error(fmt.Sprintf("Edge %s (or its reverse) is duplicated.", e))
			}
			edgeSet[e] = true
		}
	}
	return edgeSet, nil
}

/*** OutFileReader ***/

func NewOutFileReader(rd io.Reader) *OutFileReader {
	var outReader OutFileReader
	outReader.reader = bufio.NewReader(rd)
	return &outReader
}

func (r *OutFileReader) ReadOutputFile(edgeSets []EdgeSet) ([]int, *Error) {
	nums, err := r.readNumbers("Cannot parse the number of output graphs.", 1)
	if err != nil {
		return nil, err
	}

	if nums[0] != len(edgeSets) {
		return nil, r.Error(fmt.Sprintf("The number of output graphs (%d) should "+
			"equal the number of input graphs (%d).", nums[0], len(edgeSets)))
	}

	NumCases := len(edgeSets)
	NumLeaves := make([]int, NumCases)
	for r.caseNum = 1; r.caseNum <= NumCases; r.caseNum++ {
		NumLeaves[r.caseNum-1], err = r.ReadOutputGraph(edgeSets[r.caseNum-1])
		if err != nil {
			return nil, err
		}
	}

	if err := r.readLine(); err != io.EOF {
		return nil, r.ErrorWithExpected(fmt.Sprintf(
			"Extra lines after Graph #%d (line 1 says the number of graphs is %d).",
			NumCases, NumCases), "Expecting EOF.")
	}
	return NumLeaves, nil
}

func (r *OutFileReader) ReadOutputGraph(inEdgeSet EdgeSet) (int, *Error) {
	nums, err := r.readNumbers("Cannot parse the number of edges.", 1)
	if err != nil {
		return InvalidNumLeaves, err
	}

	Gin := inEdgeSet.Graph()
	Gin.Search()
	if nums[0] != Gin.NumNodes-1 {
		return InvalidNumLeaves, r.Error(fmt.Sprintf("Input graph has %d "+
			"non-isolated nodes, output graph should have %d edges, got %d instead.",
			Gin.NumNodes, Gin.NumNodes-1, nums[0]))
	}

	outEdgeSet := make(EdgeSet)
	NumEdges := nums[0]
	for i := 0; i < NumEdges; i++ {
		nums, err = r.readNumbers("Cannot parse the next edge.", 2)
		if err != nil {
			return InvalidNumLeaves, err
		} else {
			var e Edge
			e.Ends[0] = nums[0]
			e.Ends[1] = nums[1]
			e.Normalize()

			if _, ok := inEdgeSet[e]; !ok {
				return InvalidNumLeaves, r.Error(fmt.Sprintf("Edge %s in the output "+
					"graph is absent in the input graph.", e))
			}

			if _, ok := outEdgeSet[e]; ok {
				return InvalidNumLeaves, r.Error(fmt.Sprintf(
					"Edge %s (or its reverse) is duplicated.", e))
			}
			outEdgeSet[e] = true
		}
	}

	Gout := outEdgeSet.Graph()
	Gout.Search()
	if Gout.NumNodes != Gin.NumNodes {
		return InvalidNumLeaves, r.Error(fmt.Sprintf(
			"Aftering reading the last edge, the number of "+
				"non-isolated nodes in the output graph (%d) should equal "+
				"that of the input graph (%d) to be a spanning tree.",
			Gout.NumNodes, Gin.NumNodes))
	}
	if !Gout.EdgesInOneComponent() {
		return InvalidNumLeaves, r.Error("Disconnected graph: after reading " +
			"the last edge, the output graph should be connected to be a spanning tree.")
	}
	if Gout.HasCycle {
		return InvalidNumLeaves, r.Error("Cycle detected: after reading the " +
			"last edge, the output graph should not have cycles to be a spanning tree.")
	}

	return Gout.NumLeaves, nil
}

/*** TeamInfoReader ***/

type Student struct {
	Login, Name string
}

type TeamInfo struct {
	Name    string
	Members []Student
}

func NewTeamInfoReader(rd io.Reader) *TeamInfoReader {
	var teamReader TeamInfoReader
	teamReader.reader = bufio.NewReader(rd)
	return &teamReader
}

func (r *TeamInfoReader) ReadTeamFile() (*TeamInfo, *Error) {
	var team TeamInfo
	if err := r.readLine(); err != nil {
		return nil, r.ErrorWithExpected(fmt.Sprintf("Cannot parse team name (%s).",
			err.Error()), "Expecting the team name on the first line.")
	}

	team.Name = strings.TrimSpace(r.line)
	if len(team.Name) == 0 {
		return nil, r.ErrorWithExpected("Team name cannot be blank.",
			"Expecting a non-blank team name on the first line.")
	}

	team.Members = make([]Student, 0)
	for err := r.readLine(); err != io.EOF; err = r.readLine() {
		if err != nil {
			return nil, r.Error(fmt.Sprintf("Cannot parse the next student (%s).",
				err.Error()))
		}

		split := strings.SplitN(r.line, " ", 2)
		login := strings.TrimSpace(split[0])

		if !LoginFormatRegexp.MatchString(login) {
			return nil, r.ErrorWithExpected(fmt.Sprintf("Student login (%s) is not valid.",
				login), LoginExpectedMsg)
		}

		if len(split) == 1 {
			return nil, r.ErrorWithExpected("Cannot parse student name.",
				"Student login and student name separated by a single space on a line.")
		}

		name := strings.TrimSpace(split[1])
		if len(name) == 0 {
			return nil, r.Error(fmt.Sprintf(
				"Student name should not be blank. Got (%s).", name))
		}

		student := Student{Login: login, Name: name}
		team.Members = append(team.Members, student)
	}

	if len(team.Members) == 0 {
		return &team, r.Error("There should be at least one student. Got none.")
	}

	return &team, nil
}
