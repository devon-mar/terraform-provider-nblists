package provider

import (
	"fmt"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestListDataSource(t *testing.T) {
	token := "abcdefghijklmnop"
	h := newTestListsHandler(t, token)
	h.addList("devices", map[string][]string{"test": {"1"}}, []string{"192.0.2.1"})
	h.addList("devices", map[string][]string{"test": {"2"}}, []string{})
	h.addList("devices", map[string][]string{"test": {"3"}, "as_cidr": {"true"}}, []string{"192.0.2.10/32", "2001:db8::1/128"})
	h.addList("devices", map[string][]string{"test": {"4"}, "as_cidr": {"false"}}, []string{"192.0.2.10"})
	h.addList("prefixes", nil, []string{"192.0.2.0/27", "192.0.2.32/27"})
	h.addList("ip-addresses", map[string][]string{"tag": {"5"}}, []string{"192.0.2.5"})
	h.addList("ip-addresses", map[string][]string{"tag": {"6"}}, []string{"192.0.2.6"})
	h.addList("ip-addresses", map[string][]string{"tag": {"7"}}, []string{"192.0.2.7"})
	h.addList("ip-addresses", map[string][]string{"tag": {"6", "7"}}, []string{"192.0.2.6", "192.0.2.7"})
	h.addList("ip-addresses", map[string][]string{"tag": {"8"}, "summarize": {"true"}}, []string{"192.0.2.0/31"})
	h.addList("ip-addresses", map[string][]string{"tag": {"9"}, "summarize": {"false"}}, []string{"192.0.2.0", "192.0.2.1"})
	s := httptest.NewServer(h)
	defer s.Close()

	providerConf := fmt.Sprintf(`
provider "nblists" {
    url = "%s"
	token = "%s"
}
`, s.URL, token)

	// basic - one IP
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConf + `
// One IP
data "nblists_list" "one" {
	endpoint = "devices"
	filter = { "test" = ["1"] }
}

// Empty
data "nblists_list" "two" {
	endpoint = "devices"
	filter = { "test" = ["2"] }
}

// as_cidr=True
data "nblists_list" "three" {
	endpoint = "devices"
	filter = { "test" = ["3"] }
	as_cidr = true
	split_af = true
	no_cidr_single_ip = true
}

// as_cidr=False
data "nblists_list" "four" {
	endpoint = "devices"
	filter = { "test" = ["4"] }
	as_cidr = false
}

// min
data "nblists_list" "five" {
	endpoint = "ip-addresses"
	filter = { "tag" = ["5"] }
	min = 1
}

// max
data "nblists_list" "six" {
	endpoint = "ip-addresses"
	filter = { "tag" = ["6"] }
	max = 1
}

// min and max
data "nblists_list" "seven" {
	endpoint = "ip-addresses"
	filter = { "tag" = ["7"] }
	min = 1
	max = 1
}

// summarize=true
data "nblists_list" "eight" {
	endpoint = "ip-addresses"
	filter = { "tag" = ["8"] }
	summarize = true
}
// summarize=false
data "nblists_list" "nine" {
	endpoint = "ip-addresses"
	filter = { "tag" = ["9"] }
	summarize = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.nblists_list.one",
						"list.0",
						"192.0.2.1",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.one",
						"list.#",
						"1",
					),
					resource.TestCheckNoResourceAttr(
						"data.nblists_list.one",
						"list_no_cidr",
					),

					resource.TestCheckResourceAttr(
						"data.nblists_list.two",
						"list.#",
						"0",
					),

					resource.TestCheckResourceAttr(
						"data.nblists_list.three",
						"list.0",
						"192.0.2.10/32",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.three",
						"list.1",
						"2001:db8::1/128",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.three",
						"list.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.three",
						"list_no_cidr.0",
						"192.0.2.10",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.three",
						"list_no_cidr.1",
						"2001:db8::1",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.three",
						"list_no_cidr.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.three",
						"list4.0",
						"192.0.2.10/32",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.three",
						"list4.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.three",
						"list6.0",
						"2001:db8::1/128",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.three",
						"list6.#",
						"1",
					),

					resource.TestCheckResourceAttr(
						"data.nblists_list.four",
						"list.0",
						"192.0.2.10",
					),

					resource.TestCheckResourceAttr(
						"data.nblists_list.five",
						"list.0",
						"192.0.2.5",
					),

					resource.TestCheckResourceAttr(
						"data.nblists_list.six",
						"list.0",
						"192.0.2.6",
					),

					resource.TestCheckResourceAttr(
						"data.nblists_list.seven",
						"list.0",
						"192.0.2.7",
					),

					resource.TestCheckResourceAttr(
						"data.nblists_list.eight",
						"list.0",
						"192.0.2.0/31",
					),

					resource.TestCheckResourceAttr(
						"data.nblists_list.nine",
						"list.0",
						"192.0.2.0",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.nine",
						"list.1",
						"192.0.2.1",
					),
				),
			},
		},
	})

	// with allow_empty_filter=true
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "nblists" {
    url = "%s"
	token = "%s"
	allow_empty_filter = true
}

data "nblists_list" "test" {
	endpoint = "prefixes"
}
`, s.URL, token),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.nblists_list.test",
						"list.0",
						"192.0.2.0/27",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.test",
						"list.1",
						"192.0.2.32/27",
					),
					resource.TestCheckNoResourceAttr(
						"data.nblists_list.test",
						"list.2",
					),
				),
			},
		},
	})

	// empty filter with allow_empty_filter=false
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConf + `
data "nblists_list" "test" {
	endpoint = "prefixes"
}
`,
				ExpectError: regexp.MustCompile(`Error getting list: filter is nil or empty`),
			},
		},
	})

	// min violated
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConf + `
data "nblists_list" "test" {
	endpoint = "ip-addresses"
	min = 2
	filter = {
		tag = ["6"]
	}
}
`,
				ExpectError: regexp.MustCompile(`The list has length \(1\) less than the min \(2\)`),
			},
		},
	})

	// max violated
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConf + `
data "nblists_list" "test" {
	endpoint = "ip-addresses"
	max = 1
	filter = {
		tag = ["6", "7"]
	}
}
`,
				ExpectError: regexp.MustCompile(`The list has length \(2\) greater than the max \(1\)`),
			},
		},
	})
}
