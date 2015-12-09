// #GoLit Tests
package main

import "testing"

func TestNameHasGoLit_stringCorrect(t *testing.T) {
	dirname := "hello-golit"
	ok := nameHasGoLit(dirname)
	if !ok {
		t.FailNow()
	}
}

func TestNameHasGoLit_stringIncorrect(t *testing.T) {
	dirname := "hello"
	ok := nameHasGoLit(dirname)
	if ok {
		t.FailNow()
	}
}

func TestCheckForGoLitDir(t *testing.T) {

}
