package js

import (
	"encoding/json"
	"fmt"
	"github.com/mgutz/ansi"
	"reflect"
	"testing"
)

func red(s string) string {
	return ansi.Color(s, "red")
}
func green(s string) string {
	return ansi.Color(s, "green")
}

func TestParseString(t *testing.T) {
	for _, test := range []struct {
		String string
		Result interface{}
	}{
		{
			`{}`,
			map[string]interface{}{},
		},
		{
			`[]`,
			[]interface{}{},
		},
		{
			`{"a":"b"}`,
			map[string]interface{}{"a": "b"},
		},
		{
			`{"a":1}`,
			map[string]interface{}{"a": float64(1)},
		},
		{
			"{\t\"a\"\t:\t1\r\n}",
			map[string]interface{}{"a": float64(1)},
		},
		{
			`{"a": "a\b\r\t\n\f"}`,
			map[string]interface{}{"a": `a\b\r\t\n\f`},
		},
		{
			`{"a":"ab'c"}`,
			map[string]interface{}{"a": `ab'c`},
		},
		{
			`{"a":3.141E-10}`,
			map[string]interface{}{"a": float64(3.141e-10)},
		},
		{
			`{"a":3.141e-10}`,
			map[string]interface{}{"a": float64(3.141e-10)},
		},
		{
			`{"a":12345123456789}`,
			map[string]interface{}{"a": float64(12345123456789)},
		},
		{
			`{"a":123456789123456789123456789}`,
			map[string]interface{}{"a": float64(123456789123456789123456789)},
		},
		{
			`{"a":1.7976931348623157E308}`,
			map[string]interface{}{"a": float64(1.7976931348623157e308)},
		},
		{
			`[1,2,3,4]`,
			[]interface{}{float64(1), float64(2), float64(3), float64(4)},
		},
		{
			`["1","2","3","4"]`,
			[]interface{}{`1`, `2`, `3`, `4`},
		},
		{
			`[{}, { }, [], [ ]]`,
			[]interface{}{
				make(map[string]interface{}),
				make(map[string]interface{}),
				make([]interface{}, 0),
				make([]interface{}, 0),
			},
		},
		{
			`{"a":"\u2000\u20ff"}`,
			map[string]interface{}{"a": "\u2000\u20ff"},
		},
		{
			`{"a":"\u2000\u20FF"}`,
			map[string]interface{}{"a": "\u2000\u20ff"},
		},
		{
			`{"a":"foo://bar"}`,
			map[string]interface{}{"a": `foo://bar`},
		},
		{
			`{"\uafaf":"\uafaf"}`,
			map[string]interface{}{"꾯": "꾯"},
		},
		{
			`{"a":null}`,
			map[string]interface{}{"a": nil},
		},
		{
			`{"a":true}`,
			map[string]interface{}{"a": true},
		},
		{
			`{"a":false}`,
			map[string]interface{}{"a": false},
		},
		{
			`{"a":{"b":["c", 1]}}`,
			map[string]interface{}{"a": map[string]interface{}{"b": []interface{}{"c", float64(1)}}},
		},
	} {
		title := fmt.Sprintf("Should support parsing %q", test.String)
		parsed, err := Parse(test.String)
		if err != nil {
			t.Error(red("✘ "+title) + "\n\n" + red(fmt.Sprintf("Could not parse %q: %s", test.String, err)) + "\n")
			continue
		}

		if !reflect.DeepEqual(parsed, test.Result) {
			message := "Expectation failed:"
			expectation := fmt.Sprintf("expected: <%T> %#v", test.Result, test.Result)
			actual := fmt.Sprintf("actual:   <%T> %#v", parsed, parsed)

			t.Error(red("✘ "+title) + "\n\n" + red(message) + "\n\t" + red(expectation) + "\n\t" + red(actual) + "\n\t")
		} else {
			t.Log(green("✔ " + title))
		}
	}
}

var BigJson = `{"web-app":{"servlet":[{"servlet-name":"cofaxCDS","servlet-class":"org.cofax.cds.CDSServlet","init-param":{"configGlossary:installationAt":"Philadelphia, PA","configGlossary:adminEmail":"ksm@pobox.com","configGlossary:poweredBy":"Cofax","configGlossary:poweredByIcon":"/images/cofax.gif","configGlossary:staticPath":"/content/static","templateProcessorClass":"org.cofax.WysiwygTemplate","templateLoaderClass":"org.cofax.FilesTemplateLoader","templatePath":"templates","templateOverridePath":"","defaultListTemplate":"listTemplate.htm","defaultFileTemplate":"articleTemplate.htm","useJSP":false,"jspListTemplate":"listTemplate.jsp","jspFileTemplate":"articleTemplate.jsp","cachePackageTagsTrack":200,"cachePackageTagsStore":200,"cachePackageTagsRefresh":60,"cacheTemplatesTrack":100,"cacheTemplatesStore":50,"cacheTemplatesRefresh":15,"cachePagesTrack":200,"cachePagesStore":100,"cachePagesRefresh":10,"cachePagesDirtyRead":10,"searchEngineListTemplate":"forSearchEnginesList.htm","searchEngineFileTemplate":"forSearchEngines.htm","searchEngineRobotsDb":"WEB-INF/robots.db","useDataStore":true,"dataStoreClass":"org.cofax.SqlDataStore","redirectionClass":"org.cofax.SqlRedirection","dataStoreName":"cofax","dataStoreDriver":"com.microsoft.jdbc.sqlserver.SQLServerDriver","dataStoreUrl":"jdbc:microsoft:sqlserver://LOCALHOST:1433;DatabaseName=goon","dataStoreUser":"sa","dataStorePassword":"dataStoreTestQuery","dataStoreTestQuery":"SET NOCOUNT ON;select test='test';","dataStoreLogFile":"/usr/local/tomcat/logs/datastore.log","dataStoreInitConns":10,"dataStoreMaxConns":100,"dataStoreConnUsageLimit":100,"dataStoreLogLevel":"debug","maxUrlLength":500}},{"servlet-name":"cofaxEmail","servlet-class":"org.cofax.cds.EmailServlet","init-param":{"mailHost":"mail1","mailHostOverride":"mail2"}},{"servlet-name":"cofaxAdmin","servlet-class":"org.cofax.cds.AdminServlet"},{"servlet-name":"fileServlet","servlet-class":"org.cofax.cds.FileServlet"},{"servlet-name":"cofaxTools","servlet-class":"org.cofax.cms.CofaxToolsServlet","init-param":{"templatePath":"toolstemplates/","log":1,"logLocation":"/usr/local/tomcat/logs/CofaxTools.log","logMaxSize":"","dataLog":1,"dataLogLocation":"/usr/local/tomcat/logs/dataLog.log","dataLogMaxSize":"","removePageCache":"/content/admin/remove?cache=pages&id=","removeTemplateCache":"/content/admin/remove?cache=templates&id=","fileTransferFolder":"/usr/local/tomcat/webapps/content/fileTransferFolder","lookInContext":1,"adminGroupID":4,"betaServer":true}}],"servlet-mapping":{"cofaxCDS":"/","cofaxEmail":"/cofaxutil/aemail/*","cofaxAdmin":"/admin/*","fileServlet":"/static/*","cofaxTools":"/tools/*"},"taglib":{"taglib-uri":"cofax.tld","taglib-location":"/WEB-INF/tlds/cofax.tld"}}}`
var SmallJson = `{"a":"b", "c":[1,2,3], "d":{}}`

func BenchmarkBig(t *testing.B) {
	for i := 0; i < t.N; i++ {
		if _, err := Parse(BigJson); err != nil {
			t.Fatal(err)
		}
	}
}
func BenchmarkSmall(t *testing.B) {
	for i := 0; i < t.N; i++ {
		if _, err := Parse(SmallJson); err != nil {
			t.Fatal(err)
		}
	}
}
func BenchmarkBigNative(t *testing.B) {
	for i := 0; i < t.N; i++ {
		var to interface{}
		if err := json.Unmarshal([]byte(BigJson), to); err != nil {
			t.Fatal(err)
		}
	}
}
func BenchmarkSmallNative(t *testing.B) {
	for i := 0; i < t.N; i++ {
		var to interface{}
		if err := json.Unmarshal([]byte(SmallJson), to); err != nil {
			t.Fatal(err)
		}
	}
}
