package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/CTSRD-SOAAP/gosoaap"
)

type Analyses []string

func (a *Analyses) Set(value string) error {
	*a = strings.Split(value, ",")
	return nil
}

func (a Analyses) String() string {
	return strings.Join(a, ", ")
}

func main() {
	//
	// Command-line arguments:
	//
	analyses := Analyses{"vuln"}
	flag.Var(&analyses, "analyses",
		"SOAAP analysis results to graph (options: "+
			strings.Join(soaap.GraphAnalyses(), ", ")+
			")")

	output := flag.String("output", "-", "output file")

	binout := flag.Bool("binary", false, "write binary output")

	flag.Parse()

	if len(flag.Args()) != 1 {
		printUsage()
		return
	}

	input := flag.Args()[0]

	//
	// Load input file:
	//
	f, err := os.Open(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		return
	}

	//
	// Open output file:
	//
	var out *os.File
	if *output == "-" {
		out = os.Stdout
	} else {
		out, err = os.Create(*output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			return
		}
	}

	//
	// Is the input file a binary-encoded graph or a set of SOAAP results
	// that we need to extract a graph from?
	//
	var graph soaap.CallGraph

	if strings.HasSuffix(f.Name(), ".graph") {
		graph, err = soaap.LoadGraph(f, report)

	} else {
		graph, err = analyzeResultsFile(f, analyses)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		return
	}

	//
	// Output the results:
	//
	if *binout {
	} else {
		writeDot(graph, out)
	}
}

func dotHeader() string {
	return `

	node [ fontname = "Inconsolata" ];
	edge [ fontname = "Avenir" ];

	labeljust = "l";
	labelloc = "b";
	rankdir = "BT";

`
}

func analyzeResultsFile(f *os.File, analyses []string) (soaap.CallGraph, error) {
	//
	// Combine callgraphs of the requested analyses:
	//
	results, err := soaap.LoadResults(f, report)
	if err != nil {
		return soaap.CallGraph{}, err
	}

	graph := soaap.NewCallGraph()
	for _, a := range analyses {
		g, err := results.ExtractGraph(a, report)
		if err != nil {
			return graph, err
		}

		graph.Union(g)
	}

	return graph, err
}

func printUsage() {
	fmt.Fprintf(os.Stderr,
		"Usage:  soaap-graph [options] <input file>\n\n")

	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
}

func report(progress string) {
	fmt.Println(progress)
	os.Stdout.Sync()
}

func writeDot(graph soaap.CallGraph, out *os.File) {
	fmt.Fprintln(out, "digraph {")
	fmt.Fprintln(out, dotHeader())

	for _, n := range graph.Nodes {
		fmt.Fprintf(out, "	%s\n", n.Dot())
	}

	for _, c := range graph.Calls {
		caller := graph.Nodes[c.Caller]
		callee := graph.Nodes[c.Callee]

		fmt.Fprintf(out, "	\"%s\" -> \"%s\";\n",
			caller.Name, callee.Name)
	}

	fmt.Fprintf(out, "}\n")
}
