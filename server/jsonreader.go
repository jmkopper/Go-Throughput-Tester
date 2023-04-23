import (
	"io"
	"io/ioutil"
)

func testRead(filename io.Reader) Test {
	var test Test
	byteValue, _ := ioutil.ReadAll(filename)
	json.Unmarshal(byteValue, &test)

	return test
}