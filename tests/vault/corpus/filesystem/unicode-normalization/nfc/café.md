# café (NFC)

Filename is NFC-normalized Unicode: a single codepoint for e-acute.
Lives in a directory separate from the NFD variant because this host's filesystem (APFS) treats NFC and NFD forms of the same name as the same directory entry within one directory -- writing both as siblings silently overwrote one with the other's content during corpus generation. See unicode-normalization/README.md.
