package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_PP_PrintNamesAndType(t *testing.T) {
	data := slicePP{}.PrintNamesAndTypes("DHDH", ",", false)
	assert.Equal(t, "", data, "PrintCallParams not work correctly on empty data")
	data = slicePP{
		{ParamName: "test"},
		{ParamName: "test2"},
	}.PrintNamesAndTypes("DHDH", ",", false)
	assert.Equal(t, "DHDHtest,DHDHtest2", data, "PrintCallParams not work correctly on empty data")
	data = slicePP{
		{ParamName: "test"},
		{ParamName: "test2"},
	}.PrintNamesAndTypes("DHDH", " A ", false)
	assert.Equal(t, "DHDHtest A DHDHtest2", data, "PrintCallParams not work correctly on empty data")
	data = slicePP{
		{ParamName: "test", ParamType: "qqa"},
		{ParamName: "test2", ParamType: "wer"},
	}.PrintNamesAndTypes("DHDH", " B ", true)
	assert.Equal(t, "DHDHtest qqa B DHDHtest2 wer", data, "PrintCallParams not work correctly on empty data")
}

func Test_PA_PrintNamesAndType(t *testing.T) {
	data := slicePA{}.PrintNamesAndTypes("DHDH", ",", false)
	assert.Equal(t, "", data, "PrintCallParams not work correctly on empty data")
	data = slicePA{
		{ArgName: "test"},
		{ArgName: "test2"},
	}.PrintNamesAndTypes("DHDH", ",", false)
	assert.Equal(t, "DHDHtest,DHDHtest2", data, "PrintCallParams not work correctly on empty data")
	data = slicePA{
		{ArgName: "test"},
		{ArgName: "test2"},
	}.PrintNamesAndTypes("DHDH", " A ", false)
	assert.Equal(t, "DHDHtest A DHDHtest2", data, "PrintCallParams not work correctly on empty data")
	data = slicePA{
		{ArgName: "test", ArgType: "qqa"},
		{ArgName: "test2", ArgType: "wer"},
	}.PrintNamesAndTypes("DHDH", " B ", true)
	assert.Equal(t, "DHDHtest qqa B DHDHtest2 wer", data, "PrintCallParams not work correctly on empty data")
	data = slicePA{
		{ArgName: "test", ArgType: "qqa"},
	}.PrintNamesAndTypes("DHDH", " B ", true)
	assert.Equal(t, "DHDHtest qqa", data, "PrintCallParams not work correctly on empty data")

}
