import pynetbox
import os
from typing import Dict, List
from argparse import ArgumentParser


def create_ips(nb: pynetbox.api):
    tag_ips: Dict[str, List[str]] = {
        "1": ["192.0.2.1/24"],
        "2": [],
        "3": ["192.0.2.3/27", "2001:db8::3/64"],
        "4": ["192.0.2.4/27"],
        "5": ["192.0.2.5/27"],
        "6": ["192.0.2.6/27"],
        "7": ["192.0.2.7/27"],
        "8": ["192.0.2.100/31"],
        "9": ["192.0.2.101/31"],
        "10": ["192.0.2.10/27", "192.0.2.11/27"],
        "11": ["192.0.2.12/27", "2001:db8::12/64"],
    }
    for tag, ips in tag_ips.items():
        tag = nb.extras.tags.create(name=tag, slug=tag)
        for ip in ips:
            nb.ipam.ip_addresses.create(address=ip, tags=[tag.id])


def create_aggregates(nb: pynetbox.api):
    rir = nb.ipam.rirs.create(name="RFC3330", slug="rfc3330")
    prefixes: List[str] = ["192.0.2.0/24"]
    for p in prefixes:
        nb.ipam.aggregates.create(prefix=p, rir=rir.id)


def create_prefixes(nb: pynetbox.api):
    tag_prefixes: Dict[str, List[str]] = {
        "p1": [
            "192.0.2.0/27",
            "192.0.2.200/32",
            "2001:db8::/64",
            "2001:db8::200/128",
        ]
    }
    for tag, prefixes in tag_prefixes.items():
        tag = nb.extras.tags.create(name=tag, slug=tag)
        for p in prefixes:
            nb.ipam.prefixes.create(prefix=p, tags=[tag.id])


def main():
    parser = ArgumentParser()
    parser.add_argument(
        "--delete", action="store_true", help="Delete all prefixes, IPs and tags."
    )
    args = parser.parse_args()

    nb = pynetbox.api(
        os.environ["TEST_NBLISTS_URL"], token=os.environ["TEST_NBLISTS_TOKEN"]
    )
    if args.delete:
        nb.ipam.ip_addresses.delete(nb.ipam.ip_addresses.all())
        nb.ipam.prefixes.delete(nb.ipam.prefixes.all())
        nb.ipam.aggregates.delete(nb.ipam.aggregates.all())
        nb.ipam.rirs.delete(nb.ipam.rirs.all())
        nb.extras.tags.delete(nb.extras.tags.all())
    else:
        create_ips(nb)
        create_prefixes(nb)
        create_aggregates(nb)


if __name__ == "__main__":
    main()
