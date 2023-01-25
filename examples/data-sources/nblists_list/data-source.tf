data "nblists_list" "special_ips" {
  endpoint = "ip-addresses"
  filter = {
    tag = ["special"]
  }
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
