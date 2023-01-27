data "nblists_list" "special" {
  endpoint = "ip-addresses"
  filter = {
    tag = ["special"]
  }
  no_cidr_single_ip = true
}

data "nblists_list" "vm" {
  endpoint = "virtual-machines"
  filter = {
    name = ["VM01"]
  }
  // Populate list4 and list6.
  // In this case, `list4` and `list6` will contain the device's primary
  // IPv4 and IPv6 address respectively (if any).
  // Both IP addresses of the VM can also be accessed with `list` as well.
  split_af = true
}

# Use the data
resource "some_resource" "r" {
  cidrs = data.nblists_list.special.list

  // If idempotency doesn't work with '/32' or '/128' prefixes
  // we an set `no_cidr_single_ip=True` and use the `list_no_cidr` attribute.
  // This way, we can both representations but only make one request to NetBox.
  // 
  // We could also just set `as_cidr=False`
  hosts = data.nblists_list.special.list_no_cidr
}
