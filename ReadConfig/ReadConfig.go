package ReadConfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"

	"github.com/fatih/structtag"
	"github.com/pschlump/jsonSyntaxErrorLib"
)

//// TODO - TODO - add in recursive setting - xyzzy0001
/*
	case reflect.Struct:
            walk(field.Interface(), fn)
	https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/reflection
		-- Has code for walk of Strtuct/Array/Map etc.
		-- Would need "set" capability
	https://medium.freecodecamp.org/a-practical-example-go-reflections-and-generic-designs-4868b6cdb2dc
	https://gist.github.com/tkrajina/880eb4b9a10aee28707e2aa764257503

	1. Validation of Fields (stings/dates etc)
	2. Default values
	3. Pull in ENV
	4. $PROMPT$ - pull in password from STDIN/quite
	5. Change $ENV$ and $FILE$ into "tag" `env:"EnvName"` `file:"FileName"` `prompt:"PromptTag","password"`
	6. Use of tags allow setting of int/int32/int64 etc.

	1. Create a "ConfigureProgram" package.
*/

// ReadFile will read a configuration file into the global configuration structure.
func ReadFile(filename string, lCfg interface{}) (err error) {

	// Get the type and value of the argument we were passed.
	ptyp := reflect.TypeOf(lCfg)
	pval := reflect.ValueOf(lCfg)

	// Requries that lCfg is a pointer.
	if ptyp.Kind() != reflect.Ptr {
		return fmt.Errorf("Must pass a address of a struct to RedFile")
	}

	var typ reflect.Type
	var val reflect.Value
	typ = ptyp.Elem()
	val = pval.Elem()

	// Create Defaults

	// Make sure we now have a struct
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("ReadFile was not passed a struct.")
	}

	// Can we set values?
	if val.CanSet() {
		if db1 {
			fmt.Printf("Debug: We can set values.\n")
		}
	} else {
		return fmt.Errorf("ReadFile passed a struct that will not allow setting of values")
	}

	// The number of fields in the struct is determined by the type of struct it is. Loop through them.
	for i := 0; i < typ.NumField(); i++ {

		// Get the type of the field from the type of the struct. For a struct, you always get a StructField.
		sfld := typ.Field(i)

		// Get the type of the StructField, which is the type actually stored in that field of the struct.
		tfld := sfld.Type

		// Get the Kind of that type, which will be the underlying base type
		// used to define the type in question.
		kind := tfld.Kind()

		// Get the value of the field from the value of the struct.
		vfld := val.Field(i)
		tag := string(sfld.Tag)

		// ... and start using structtag by parsing the tag
		tags, err := structtag.Parse(tag)
		if err != nil {
			return fmt.Errorf("Unable to parse structure tag ->%s<- %s", tag, err)
		}

		// Dump out what we've found
		if db1 {
			fmt.Printf("Debug: struct field %d: name %s type %s kind %s value %v tag ->%s<-\n", i, sfld.Name, tfld, kind, vfld, tag)

			// iterate over all tags
			for tn, t := range tags.Tags() {
				fmt.Printf("\t[%d] tag: %+v\n", tn, t)
			}

			// get a single tag
			defaultTag, err := tags.Get("default")
			if err != nil {
				fmt.Printf("`default` Not Set\n")
			} else {
				fmt.Println(defaultTag)         // Output: default:"foo,omitempty,string"
				fmt.Println(defaultTag.Key)     // Output: default
				fmt.Println(defaultTag.Name)    // Output: foo
				fmt.Println(defaultTag.Options) // Output: [omitempty string]
			}
		}

		defaultTag, err := tags.Get("default")
		// Is that field some kind of string, and is the value one we can set?
		// 1. Other tyeps (all ints, floats) - not just strings		xyzzy001-type
		if kind == reflect.String && vfld.CanSet() {
			if err != nil || defaultTag.Name == "" {
				// Ignore error - indicates no "default" tag set.
			} else {
				defaultValue := defaultTag.Name
				if db1 {
					fmt.Printf("Debug: Looking to set field %s to a default value of ->%s<-\n", sfld.Name, defaultValue)
				}
				vfld.SetString(defaultValue)
			}
		} else if kind != reflect.String && err == nil {
			// report errors - defauilt is only implemented with strings.
			return fmt.Errorf("default tag on struct is only implemented for String fields that are settable in struct.  Fatal error on %s tag %s", sfld.Name, tag)
		}
		// TODO - TODO - other types - xyzzy0001
		// } else if kind == reflect.Struct && ...
		// } else if kind == Pointers?
		// } else if kind == Maps?
		// } else if kind == Array?
		// } else if kind == Slice?
	}

	// look for filename in ~/local (C:\local on Winderz)
	var home string
	// If we are on Windows then pull from c:/, else use the home direcotry.
	if os.PathSeparator == '/' {
		home = os.Getenv("HOME")
	} else {
		home = "C:\\"
	}
	homeLocal := path.Join(home, "local")
	base := path.Base(filename)
	if ExistsIsDir(homeLocal) && Exists(path.Join(homeLocal, base)) {
		filename = path.Join(homeLocal, base)
	}
	if db1 {
		fmt.Printf("Debug: File name after checing ~/local [%s]\n", filename)
	}

	var buf []byte
	buf, err = ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Unable to read the JSON file [%s]: error %s", filename, err)
	}

	// err = json.Unmarshal(buf, &gCfg)
	err = json.Unmarshal(buf, lCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid initialization - Unable to parse JSON file, %s\n", err)
		msg := PrintErrorJson(string(buf), err) // show line for error
		return fmt.Errorf("Invalid initialization - Unable to parse JSON file, %s %s", err, msg)
	}

	err = setFromEnv2(typ, val)
	if err != nil {
		return fmt.Errorf("Error pulling from environment: %s", err)
	}

	return err
}

func PrintErrorJson(js string, err error) (rv string) {
	rv = jsonSyntaxErrorLib.GenerateSyntaxError(js, err)
	fmt.Fprintf(os.Stderr, "%s\n", rv)
	return
}

func setFromEnv2(typ reflect.Type, val reflect.Value) (err error) {

	// The number of fields in the struct is determined by the type of struct
	// it is. Loop through them.
	for i := 0; i < typ.NumField(); i++ {

		// Get the type of the field from the type of the struct. For a struct, you always get a StructField.
		sfld := typ.Field(i)

		// Get the type of the StructField, which is the type actually stored in that field of the struct.
		tfld := sfld.Type

		// Get the Kind of that type, which will be the underlying base type
		// used to define the type in question.
		kind := tfld.Kind()

		// Get the value of the field from the value of the struct.
		vfld := val.Field(i)

		// Dump out what we've found
		if db2 {
			fmt.Printf("Debug: struct field %d: name %s type %s kind %s value %v\n", i, sfld.Name, tfld, kind, vfld)
		}

		// Is that field some kind of string, and is the value one we can set?
		if kind == reflect.String && vfld.CanSet() {
			if db2 {
				fmt.Printf("Debug: Looking to set field %s\n", sfld.Name)
			}
			// Assign to it
			curVal := fmt.Sprintf("%s", vfld)
			if len(curVal) > 5 && curVal[0:5] == "$ENV$" {
				envVal := os.Getenv(curVal[5:])
				if db2 {
					fmt.Printf("Debug: Overwriting field %s current [%s] with [%s]\n", sfld.Name, curVal, envVal)
				}
				vfld.SetString(envVal)
			}
			if len(curVal) > 6 && curVal[0:6] == "$FILE$" {
				data, err := ioutil.ReadFile(curVal[6:])
				if db2 {
					fmt.Printf("Debug: Overwriting field %s current [%s] with [%s]\n", sfld.Name, data, data)
				}
				if err != nil {
					return fmt.Errorf("Error [%s] with file [%s] field name [%s]", err, curVal[6:], sfld.Name)
				}
				vfld.SetString(string(data))
			}
		}
		//// TODO - TODO - add in recursive setting - xyzzy0001
		// } else if kind == reflect.Struct && ...
		// } else if kind == Pointers?
		// } else if kind == Maps?
		// } else if kind == Array?
		// } else if kind == Slice?
	}

	return nil
}

func SetFromEnv(s interface{}) (err error) {

	// Get the type and value of the argument we were passed.
	ptyp := reflect.TypeOf(s)
	pval := reflect.ValueOf(s)
	// We can't do much with the Value (it's opaque), but we need it in order
	// to fetch individual fields from the struct later.

	var typ reflect.Type
	var val reflect.Value

	// If we were passed a pointer, dereference to get the type and value
	// pointed at.
	if ptyp.Kind() == reflect.Ptr {
		if db2 {
			fmt.Printf("Debug: Argument is a pointer, dereferencing.\n")
		}
		typ = ptyp.Elem()
		val = pval.Elem()
	} else {
		if db2 {
			fmt.Printf("Debug: Argument is %s.%s, a %s.\n", ptyp.PkgPath(), ptyp.Name(), ptyp.Kind())
		}
		typ = ptyp
		val = pval
	}

	// Make sure we now have a struct
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("SetFromEnv was not passed a struct.")
	}

	// Can we set values?
	if val.CanSet() {
		if db2 {
			fmt.Printf("Debug: We can set values.\n")
		}
	} else {
		return fmt.Errorf("SetFromEnv passed a struct that will not allow setting of values")
	}

	return setFromEnv2(typ, val)

}

// Exists returns true if a directory or file exists.
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// ExistsIsDir returns true if a direcotry exists.
func ExistsIsDir(name string) bool {
	fi, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	if fi.IsDir() {
		return true
	}
	return false
}

var db1 = false
var db2 = false
