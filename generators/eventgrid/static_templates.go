// generated by github.com/Azure/buffalo-azure/generators/eventgrid/builder
// DO NOT EDIT - instead run "go generate ./..."

package eventgrid

var staticTemplates = make(TemplateCache)

func init() {
	staticTemplates["templates/actions/eventgrid_name.go"] = []byte{112, 97, 99, 107, 97, 103, 101, 32, 97, 99, 116, 105, 111, 110, 115, 10, 10, 105, 109, 112, 111, 114, 116, 32, 40, 10, 9, 34, 101, 110, 99, 111, 100, 105, 110, 103, 47, 106, 115, 111, 110, 34, 10, 9, 34, 101, 114, 114, 111, 114, 115, 34, 10, 9, 34, 110, 101, 116, 47, 104, 116, 116, 112, 34, 10, 10, 9, 34, 103, 105, 116, 104, 117, 98, 46, 99, 111, 109, 47, 65, 122, 117, 114, 101, 47, 98, 117, 102, 102, 97, 108, 111, 45, 97, 122, 117, 114, 101, 47, 115, 100, 107, 47, 101, 118, 101, 110, 116, 103, 114, 105, 100, 34, 10, 9, 34, 103, 105, 116, 104, 117, 98, 46, 99, 111, 109, 47, 103, 111, 98, 117, 102, 102, 97, 108, 111, 47, 98, 117, 102, 102, 97, 108, 111, 34, 10, 41, 10, 10, 47, 47, 32, 77, 121, 69, 118, 101, 110, 116, 71, 114, 105, 100, 84, 111, 112, 105, 99, 83, 117, 98, 115, 99, 114, 105, 98, 101, 114, 32, 103, 97, 116, 104, 101, 114, 115, 32, 114, 101, 115, 112, 111, 110, 100, 115, 32, 116, 111, 32, 97, 108, 108, 32, 82, 101, 113, 117, 101, 115, 116, 115, 32, 115, 101, 110, 116, 32, 116, 111, 32, 97, 32, 112, 97, 114, 116, 105, 99, 117, 108, 97, 114, 32, 101, 110, 100, 112, 111, 105, 110, 116, 46, 10, 116, 121, 112, 101, 32, 77, 121, 69, 118, 101, 110, 116, 71, 114, 105, 100, 84, 111, 112, 105, 99, 83, 117, 98, 115, 99, 114, 105, 98, 101, 114, 32, 115, 116, 114, 117, 99, 116, 32, 123, 10, 9, 101, 118, 101, 110, 116, 103, 114, 105, 100, 46, 83, 117, 98, 115, 99, 114, 105, 98, 101, 114, 10, 125, 10, 10, 47, 47, 32, 78, 101, 119, 77, 121, 69, 118, 101, 110, 116, 71, 114, 105, 100, 84, 111, 112, 105, 99, 83, 117, 98, 115, 99, 114, 105, 98, 101, 114, 32, 105, 110, 115, 116, 97, 110, 116, 105, 97, 116, 101, 115, 32, 77, 121, 69, 118, 101, 110, 116, 71, 114, 105, 100, 84, 111, 112, 105, 99, 83, 117, 98, 115, 99, 114, 105, 98, 101, 114, 32, 102, 111, 114, 32, 117, 115, 101, 32, 105, 110, 32, 97, 32, 96, 98, 117, 102, 102, 97, 108, 111, 46, 65, 112, 112, 96, 46, 10, 102, 117, 110, 99, 32, 78, 101, 119, 77, 121, 69, 118, 101, 110, 116, 71, 114, 105, 100, 84, 111, 112, 105, 99, 83, 117, 98, 115, 99, 114, 105, 98, 101, 114, 40, 112, 97, 114, 101, 110, 116, 32, 101, 118, 101, 110, 116, 103, 114, 105, 100, 46, 83, 117, 98, 115, 99, 114, 105, 98, 101, 114, 41, 32, 40, 99, 114, 101, 97, 116, 101, 100, 32, 42, 77, 121, 69, 118, 101, 110, 116, 71, 114, 105, 100, 84, 111, 112, 105, 99, 83, 117, 98, 115, 99, 114, 105, 98, 101, 114, 41, 32, 123, 10, 9, 100, 105, 115, 112, 97, 116, 99, 104, 101, 114, 32, 58, 61, 32, 101, 118, 101, 110, 116, 103, 114, 105, 100, 46, 78, 101, 119, 84, 121, 112, 101, 68, 105, 115, 112, 97, 116, 99, 104, 83, 117, 98, 115, 99, 114, 105, 98, 101, 114, 40, 112, 97, 114, 101, 110, 116, 41, 10, 10, 9, 99, 114, 101, 97, 116, 101, 100, 32, 61, 32, 38, 77, 121, 69, 118, 101, 110, 116, 71, 114, 105, 100, 84, 111, 112, 105, 99, 83, 117, 98, 115, 99, 114, 105, 98, 101, 114, 123, 10, 9, 9, 83, 117, 98, 115, 99, 114, 105, 98, 101, 114, 58, 32, 100, 105, 115, 112, 97, 116, 99, 104, 101, 114, 44, 10, 9, 125, 10, 10, 9, 114, 101, 116, 117, 114, 110, 10, 125, 10, 10, 47, 47, 32, 82, 101, 99, 101, 105, 118, 101, 77, 121, 84, 121, 112, 101, 32, 119, 105, 108, 108, 32, 114, 101, 115, 112, 111, 110, 100, 32, 116, 111, 32, 97, 110, 32, 96, 101, 118, 101, 110, 116, 103, 114, 105, 100, 46, 69, 118, 101, 110, 116, 96, 32, 99, 97, 114, 114, 121, 105, 110, 103, 32, 97, 32, 115, 101, 114, 105, 97, 108, 105, 122, 101, 100, 32, 96, 77, 121, 84, 121, 112, 101, 96, 32, 97, 115, 32, 105, 116, 115, 32, 112, 97, 121, 108, 111, 97, 100, 46, 10, 102, 117, 110, 99, 32, 40, 115, 32, 42, 77, 121, 69, 118, 101, 110, 116, 71, 114, 105, 100, 84, 111, 112, 105, 99, 83, 117, 98, 115, 99, 114, 105, 98, 101, 114, 41, 32, 82, 101, 99, 101, 105, 118, 101, 77, 121, 84, 121, 112, 101, 40, 99, 32, 98, 117, 102, 102, 97, 108, 111, 46, 67, 111, 110, 116, 101, 120, 116, 44, 32, 101, 32, 101, 118, 101, 110, 116, 103, 114, 105, 100, 46, 69, 118, 101, 110, 116, 41, 32, 101, 114, 114, 111, 114, 32, 123, 10, 9, 118, 97, 114, 32, 112, 97, 121, 108, 111, 97, 100, 32, 77, 121, 84, 121, 112, 101, 10, 9, 105, 102, 32, 101, 114, 114, 32, 58, 61, 32, 106, 115, 111, 110, 46, 85, 110, 109, 97, 114, 115, 104, 97, 108, 40, 101, 46, 68, 97, 116, 97, 44, 32, 38, 112, 97, 121, 108, 111, 97, 100, 41, 59, 32, 101, 114, 114, 32, 33, 61, 32, 110, 105, 108, 32, 123, 10, 9, 9, 114, 101, 116, 117, 114, 110, 32, 99, 46, 69, 114, 114, 111, 114, 40, 104, 116, 116, 112, 46, 83, 116, 97, 116, 117, 115, 66, 97, 100, 82, 101, 113, 117, 101, 115, 116, 44, 32, 101, 114, 114, 111, 114, 115, 46, 78, 101, 119, 40, 34, 117, 110, 97, 98, 108, 101, 32, 116, 111, 32, 117, 110, 109, 97, 114, 115, 104, 97, 108, 32, 114, 101, 113, 117, 101, 115, 116, 32, 100, 97, 116, 97, 34, 41, 41, 10, 9, 125, 10, 10, 9, 47, 47, 32, 82, 101, 112, 108, 97, 99, 101, 32, 116, 104, 101, 32, 99, 111, 100, 101, 32, 98, 101, 108, 111, 119, 32, 119, 105, 116, 104, 32, 121, 111, 117, 114, 32, 108, 111, 103, 105, 99, 10, 9, 114, 101, 116, 117, 114, 110, 32, 110, 105, 108, 10, 125, 10}
	staticTemplates["templates/actions/mytype.go"] = []byte{112, 97, 99, 107, 97, 103, 101, 32, 97, 99, 116, 105, 111, 110, 115, 10, 10, 47, 47, 32, 77, 121, 84, 121, 112, 101, 32, 101, 120, 105, 115, 116, 115, 32, 111, 110, 108, 121, 32, 102, 111, 114, 32, 116, 104, 101, 32, 115, 97, 107, 101, 32, 111, 102, 32, 116, 101, 109, 112, 108, 97, 116, 105, 110, 103, 44, 32, 97, 110, 100, 32, 115, 104, 111, 117, 108, 100, 32, 110, 111, 116, 32, 98, 101, 32, 117, 115, 101, 100, 46, 10, 116, 121, 112, 101, 32, 77, 121, 84, 121, 112, 101, 32, 115, 116, 114, 117, 99, 116, 32, 123, 10, 9, 70, 105, 114, 115, 116, 32, 115, 116, 114, 105, 110, 103, 32, 96, 106, 115, 111, 110, 58, 34, 102, 105, 114, 115, 116, 44, 111, 109, 105, 116, 101, 109, 112, 116, 121, 34, 96, 10, 9, 76, 97, 115, 116, 32, 32, 115, 116, 114, 105, 110, 103, 32, 96, 106, 115, 111, 110, 58, 34, 108, 97, 115, 116, 44, 111, 109, 105, 116, 101, 109, 112, 116, 121, 34, 96, 10, 125, 10}
}
