// Copyright 2020 The KCL Authors. All rights reserved.

//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"unicode"
)

var (
	flagFilename = flag.String("file", "../grammar/kcl.lark", "set lark file")
	flagOutput   = flag.String("output", "lark_token.py", "set output file")
)

func main() {
	flag.Parse()

	larkData, err := ioutil.ReadFile(*flagFilename)
	if err != nil {
		log.Fatal(err)
	}

	var buf = new(bytes.Buffer)

	fmt.Fprintln(buf, "# Copyright 2020 The KCL Authors. All rights reserved.")
	fmt.Fprintln(buf)

	fmt.Fprintln(buf, "# Auto generated by {gen_lark_token.go & kcl.lark}; DO NOT EDIT!!!")
	fmt.Fprintln(buf)
	fmt.Fprintln(buf)

	fmt.Fprintln(buf, "class LarkToken:")
	fmt.Fprintln(buf)

	var names, comments = getLarkNames(string(larkData))

	var rule_list []string
	var tok_list []string

	for _, s := range names {
		if unicode.IsLower(rune(s[0])) {
			rule_list = append(rule_list, s)
		} else {
			tok_list = append(tok_list, s)
		}
	}

	fmt.Fprintf(buf, "    # kcl.lark rules and tokens (len=%d)\n", len(names))
	for i, s := range names {
		if strings.HasPrefix(comments[i], "type: ") {
			comments[i] = strings.Replace(comments[i], "type: ", "@type: ", 1)
		}
		fmt.Fprintf(buf, "    L_%s = \"%s\"  # %s ...\n", s, s, comments[i])
	}
	fmt.Fprintln(buf)

	//fmt.Fprintf(buf, "    # Lark rule alias name (=> f'LL_{rule_name.upper()}'\n")
	//for _, s := range rule_list {
	//	fmt.Fprintf(buf, "    LL_%s = L_%s\n", strings.ToUpper(s), s)
	//}
	//fmt.Fprintln(buf)

	fmt.Fprintf(buf, "    # kcl.lark tokens list (len=%d)\n", len(tok_list))

	fmt.Fprintln(buf, "    LL_token_list = [")
	for _, s := range tok_list {
		fmt.Fprintf(buf, "        L_%s,\n", s)
	}
	fmt.Fprintln(buf, "    ]")

	fmt.Fprintln(buf)
	fmt.Fprintf(buf, "    # kcl.lark rules list (len=%d)\n", len(rule_list))

	fmt.Fprintln(buf, "    LL_rule_list = [")
	for _, s := range rule_list {
		fmt.Fprintf(buf, "        L_%s,\n", s)
	}
	fmt.Fprintln(buf, "    ]")

	fmt.Fprintln(buf)
	fmt.Fprintf(buf, "    # kcl.lark tokens string value map\n")
	fmt.Fprintln(buf, "    LL_token_str_value_map = {")
	for i, s := range names {
		if unicode.IsUpper(rune(s[0])) {
			if val := getTokenStrValue(comments[i]); val != "" {
				fmt.Fprintf(buf, "        L_%s: \"%s\",\n", s, val)
			}
		}
	}
	fmt.Fprintln(buf, "    }")

	fmt.Fprintln(buf)
	fmt.Fprintln(buf)
	fmt.Fprintln(buf, "class TokenValue:")
	for i, s := range names {
		if unicode.IsUpper(rune(s[0])) {
			if val := getTokenStrValue(comments[i]); val != "" {
				fmt.Fprintf(buf, "    %s = \"%s\"\n", s, val)
			}
		}
	}

	err = ioutil.WriteFile(*flagOutput, buf.Bytes(), 0666)
	if err != nil {
		log.Fatal(err)
	}
}

func getLarkNames(larkData string) (names []string, comments []string) {
	lines := strings.Split(larkData, "\n")
	for i, line := range lines {
		line := strings.Trim(line, "? \t")
		if matched, _ := regexp.MatchString(`^\w+(\.|:)`, line); matched {
			if idx := strings.Index(line, ":"); idx > 0 {
				line = line[:idx]
			}
			if idx := strings.Index(line, "."); idx > 0 {
				line = line[:idx]
			}
			if line != "" {
				names = append(names, line)
				comments = append(comments, lines[i])
			}
		}
	}
	return
}

func getTokenStrValue(tok_comment string) string {
	// FALSE: "False" ...
	if idx := strings.Index(tok_comment, ":"); idx >= 0 {
		tok_comment = tok_comment[idx+1:]
	}

	// IMAG_NUMBER.2: /\d+j/i | FLOAT_NUMBER "j"i ...
	tok_comment = strings.TrimSpace(tok_comment)
	if s := tok_comment; s == "" || s[0] != '"' {
		return ""
	}

	tok_comment = strings.Trim(tok_comment, `'"`)
	return tok_comment
}