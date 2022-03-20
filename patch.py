import re
import sys
from lxml import etree

RE_ADC_COMMON = re.compile("ADC_Common", re.IGNORECASE)

def patch(old_svd_path, new_svd_path):
    """
    Patch an old svd file and save to a new svd file.
    """
    root = etree.parse(old_svd_path)
    for c in root.findall("./peripherals/peripheral"):
        name_node = c.find("./name")
        name = name_node.text
        if not RE_ADC_COMMON.search(name):
            continue

        name_node.text = RE_ADC_COMMON.sub("ADC_Common", name)

        group_node = c.find("./groupName")
        if group_node is None:
            continue

        group_node.text = "ADC_Common"

    with open(new_svd_path, "wb") as f:
        f.write(etree.tostring(root))

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: {} <old_svd_path> <new_svd_path>".format(sys.argv[0]))
        sys.exit(1)
    patch(sys.argv[1], sys.argv[2])
