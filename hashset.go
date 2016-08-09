// Copyright (c) 2015, maldicion069 (Cristian Rodríguez) <ccrisrober@gmail.con>
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.package com.example

package main // set

type HashSet struct {
	data map[int32]bool
}

func (this *HashSet) Add(value int32) {
	this.data[value] = true
}

func (this *HashSet) Contains(value int32) (exists bool) {
	_, exists = this.data[value]
	return
}

// Me (13/02/2015 14:20)
func (this *HashSet) Remove(value int32) bool {
	delete(this.data, value)
	_, exists := this.data[value]
	return !exists
}

func (this *HashSet) Length() int {
	return len(this.data)
}
func (this *HashSet) RemoveDuplicates() {}

func NewSet() Set {
	return &HashSet{make(map[int32]bool)}
}

//https://github.com/karlseguin/golang-set-fun
