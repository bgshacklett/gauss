package main

import (
	"os"
	"github.com/urfave/cli"
	"io/ioutil"
	"fmt"
	"encoding/json"
	"log"
	"reflect"
	"strconv"
	"regexp"
	"github.com/beard1ess/gauss/parsing"
)

var (
	FormattedDiff parsing.Keyslice

)
var ObjectDiff = parsing.ConsumableDifference{}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func Recursion(original parsing.Keyvalue, modified parsing.Keyvalue, path parsing.Pathspec) {
	kListModified := parsing.ListStripper(modified)
	kListOriginal := parsing.ListStripper(original)
	if len(kListModified) > 1 || len(kListOriginal) > 1 {
		proc := true
		for k, v := range original {
			if parsing.IndexOf(kListModified, k) == -1 {
				removed := parsing.RemovedDifference{Path: path, Key: k, Value: v}
				ObjectDiff.Removed = append(ObjectDiff.Removed, removed)
				proc = false
			}
		}
		for k, v := range modified {
			if parsing.IndexOf(kListOriginal, k) == -1 {
				added := parsing.AddedDifference{Path: path, Key: k, Value: v}
				ObjectDiff.Added = append(ObjectDiff.Added, added)
				proc = false
			}
		}
		if proc {
			for k := range original {
				Recursion(parsing.Keyvalue{k:original[k]},parsing.Keyvalue{k:modified[k]},path)
			}
		}
		return
	}
	for k := range original {
		var npath parsing.Pathspec
		var valOrig, valMod interface{}
		if reflect.TypeOf(original).Kind() == reflect.String {
			valOrig = original
		} else {
			valOrig = original[k]
		}
		if reflect.TypeOf(modified).Kind() == reflect.String {
			valMod = modified
		} else {
			valMod = modified[k]
		}

		if !(reflect.DeepEqual(valMod, valOrig)) {
			if reflect.TypeOf(valOrig).Kind() == reflect.Map {
				npath = append(path, k)
				Recursion(parsing.Remarshal(valOrig), parsing.Remarshal(valMod), npath)
				return
			} else if reflect.TypeOf(valOrig).Kind() == reflect.Slice {
				valOrig,_ := valOrig.([]interface{})
				valMod,_ := valMod.([]interface{})
				if len(valOrig) != len(valMod) {
					// TODO array length differences
					fmt.Println("I cannot handle array length differences yet, sorry not sorry; kind of sorry.")
					os.Exit(1)
				} else {
					for i := range valOrig {
						if !(reflect.DeepEqual(valMod[i], valOrig[i])) {
							npath = append(path, "{Index:"+strconv.Itoa(i)+"}")
							
							npath = path[1]
							//, "{Index:"+strconv.Itoa(i)+"}")
							changed := parsing.ChangedDifference{Path: npath, Key: k,
								OldValue: valOrig[i], NewValue: valMod[i]}
							ObjectDiff.Changed = append(ObjectDiff.Changed, changed)
							return
						}
					}
				}
			} else {
				changed := parsing.ChangedDifference{Path: path, Key: k,
					OldValue: valOrig, NewValue: valMod}
				ObjectDiff.Changed = append(ObjectDiff.Changed, changed)
				return
			}
		}
		return
	}
	return
}


func format(input parsing.ConsumableDifference) parsing.Keyvalue {
	var return_value parsing.Keyvalue

	FormattedDiff = nil
	/*
	for i := range input["Changed"] {
		path_builder(input["Changed"][i]["Path"].([]string))
	}
	for i := range input["Added"] {
		path_builder(input["Added"][i]["Path"].([]string))
	}
	for i := range input["Removed"] {
		path_builder(input["Removed"][i]["Path"].([]string))

	}
	*/

	return return_value
}

func path_builder(path []string)  parsing.Keyvalue{
	var object parsing.Keyvalue
	FormattedDiff = nil
	r, _ := regexp.Compile("[0-9]+")
	//path_length := len(path)
	for i:= range path {
		if ok,_ := regexp.MatchString("{Index:[0-9]+}", path[i]); ok {
			index := r.FindString(path[i])
			fmt.Println(index)
		} else {

		}
	}

	fmt.Println(path)
	fmt.Println(path)
	return object
}

func main() {
	var patch, object, original_obj, modified_obj string

	app := cli.NewApp()
	app.Name = "JsonDiffer"
	app.Version = "0.1"
	app.Usage = "Used to get an object-based difference between two json objects."

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "test, t",
			Usage: "just taking up space",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "diff",
			Aliases: []string{"d"},
			Usage:   "Diff json objects",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "origin, o",
					Usage: "Original `OBJECT` to compare against",
					Value: "",
					Destination: &original_obj,
					EnvVar: "ORIGINAL_OBJECT",
				},
				cli.StringFlag{
					Name: "modified, m",
					Usage: "Modified `OBJECT` to compare against",
					Value: "",
					Destination: &modified_obj,
					EnvVar: "MODIFIED_OBJECT",
				},
				cli.StringFlag{
					Name: "output",
					Usage: "Output types available: human, machine",
					Value: "machine",
					EnvVar: "DIFF_OUTPUT",
				},
				/*
				cli.StringFlag{
					Name: "output, O",
					Usage: "File output location",
					Value: "",
					Destination: &modified_obj,
				},
				*/
			},
			Action:  func(c *cli.Context) error {
				var json_original, json_modified parsing.Keyvalue
				var path []string
				if original_obj == "" {
					fmt.Print("ORIGIN is required!\n\n")
					cli.ShowCommandHelp(c, "diff")
					os.Exit(1)
				}
				if modified_obj == "" {
					fmt.Print("MODIFIED is required!\n\n")
					cli.ShowCommandHelp(c, "diff")
					os.Exit(1)
				}

				/* TODO WE WANT TO DO ALL OUR INIT STUFF IN THIS AREA */

				/*
				ObjectDiff["Changed"] = []Keyvalue{}
				ObjectDiff["Added"] = []Keyvalue{}
				ObjectDiff["Removed"] = []Keyvalue{}
				*/

				read,err := ioutil.ReadFile(original_obj)
				check(err)
				_ = json.Unmarshal([]byte(read), &json_original)

				read,err = ioutil.ReadFile(modified_obj)
				check(err)
				_ = json.Unmarshal([]byte(read), &json_modified)


				if reflect.DeepEqual(json_original, json_modified) {
					fmt.Println("No differences!")
					os.Exit(0)
				} else {
					Recursion(json_original, json_modified, path)
				}

				if c.String("output") == "human" {
					format(ObjectDiff)
				} else if c.String("output") == "machine" {
					output,_ := json.Marshal(ObjectDiff)
					os.Stdout.Write(output)
				} else {
					fmt.Println("Output type unknown.")
					os.Exit(1)
				}

				return nil
			},
		},
		{
			Name: "patch",
			Aliases: []string{"p"},
			Usage:	"Apply patch file to json object",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "patch, p",
					Usage: "`PATCH` the OBJECT",
					Value: "",
					Destination: &patch,
				},
				cli.StringFlag{
					Name: "object, o",
					Usage: "`OBJECT` to PATCH",
					Value: "",
					Destination: &object,
				},
			},
			Action: func(c *cli.Context) error {

				return nil
			},
		},
	}

	app.Run(os.Args)

}


