// ## golit

// **golit** generates literate-programming-style HTML documentation
// from a Go source file. It produces HTML with comments alongside your
// code. Comments are parsed through [Markdown](http://daringfireball.net/projects/markdown/syntax)
// and code highlighted with [Pygments](http://pygments.org/).

// golit is based on [docco](http://jashkenas.github.com/docco/)
// and [shocco](http://rtomayko.github.com/shocco/), two earlier
// programs in the same style.

// This page is the result of running golit against its own source
// file.

package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "regexp"
    "strings"
)

// ### Usage

// golit takes exactly one argument: the path to a Go source file.
// It writes the compiled HTML on stdout.
var usage = "usage: golit input.go title > output.html"

// ### Helpers

// Panic on non-nil errors. We'll call this after error-returning
// functions.
func check(err error) {
    if err != nil {
        panic(err)
    }
}

// We'll implement Markdown rendering and Pygments syntax highlighting
// by piping the source data through external programs. This is a
// general helper for handling both cases.
func pipe(bin string, arg []string, src string) string {
    cmd := exec.Command(bin, arg...)
    in, _ := cmd.StdinPipe()
    out, _ := cmd.StdoutPipe()
    cmd.Start()
    in.Write([]byte(src))
    in.Close()
    bytes, _ := ioutil.ReadAll(out)
    err := cmd.Wait()
    check(err)
    return string(bytes)
}

// ### Rendering

// Recognize doc lines, extract their comment prefixes.
var docsPat = regexp.MustCompile("^\\s*\\/\\/\\s")

// Recognize header comment lines specially.
// var headerPat = regexp.MustCompile("^\\/\\/\\s#+\\s")

// We'll break the code into `{docs, code}` pairs, and then render
// those text segments before including them in the HTML doc.
type seg struct {
    docs, code, docsRendered, codeRendered string
}

func main() {
    // Accept exactly 2 argument, the source path and page title.
    if len(os.Args) != 3 {
        fmt.Fprintln(os.Stderr, usage)
        os.Exit(1)
    }
	sourcePath := os.Args[1]
	title := os.Args[2]

    // Ensure that we have `markdown` and `pygmentize` binaries,
    // remember their paths.
    markdownPath, err := exec.LookPath("markdown")
    check(err)
    pygmentizePath, err := exec.LookPath("pygmentize")
    check(err)

    // Read the source file in, split into lines.
    srcBytes, err := ioutil.ReadFile(sourcePath)
    check(err)
    lines := strings.Split(string(srcBytes), "\n")

    // Group lines into docs/code segments.
    segs := []*seg{}
    segs = append(segs, &seg{code: "", docs: ""})
    lastSeen := ""
    for _, line := range lines {
        lastSeg := segs[len(segs)-1]
        docsMatch := docsPat.MatchString(line)
        emptyMatch := line == ""
        lastDocs := lastSeen == "docs"
        newDocs := (lastSeen == "code") && lastSeg.docs != ""
        newCode := (lastSeen == "docs") && lastSeg.code != ""
        // Docs line - strip out comment indicator.
        if docsMatch || (emptyMatch && lastDocs) {
            trimed := docsPat.ReplaceAllString(line, "")
            if newDocs {
                newSeg := seg{docs: trimed, code: ""}
                segs = append(segs, &newSeg)
            } else {
                lastSeg.docs = lastSeg.docs + "\n" + trimed
            }
            lastSeen = "docs"
            // Code line - preserve all whitespace.
        } else {
            if newCode {
                newSeg := seg{docs: "", code: line}
                segs = append(segs, &newSeg)
            } else {
                lastSeg.code = lastSeg.code + "\n" + line
            }
            lastSeen = "code"
        }
    }

    // Render docs via `markdown` and code via `pygmentize` in each
    // segment.
    for _, seg := range segs {
        seg.docsRendered = pipe(markdownPath, []string{}, seg.docs)
        seg.codeRendered = pipe(pygmentizePath, []string{"-l", "go", "-f", "html"}, seg.code+"  ")
    }

    // Print HTML header.
    fmt.Printf(`
<!DOCTYPE html>
<html>
  <head>
    <meta http-eqiv="content-type" content="text/html;charset=utf-8">
    <title>%s</title>
    <link rel=stylesheet href="http://jashkenas.github.com/docco/resources/docco.css">
  </head>
  <body>
    <div id="container">
      <div id="background"></div>
      <table cellspacing="0" cellpadding="0">
        <thead>
          <tr>
            <td class=docs></td>
            <td class=code></td>
          </tr>
        </thead>
        <tbody>`, title)

    // Print HTML docs/code segments.
    for _, seg := range segs {
        fmt.Printf(
            `<tr>
             <td class=docs>%s</td>
             <td class=code>%s</td>
           </tr>`, seg.docsRendered, seg.codeRendered)
    }

    // Print HTML footer.
    fmt.Print(`</tbody>
           </table>
         </div>
       </body>
     </html>`)
}
