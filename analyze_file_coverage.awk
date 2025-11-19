BEGIN {
    file = ""
    total = 0
    covered = 0
}

/^github.com/ {
    # Extract filename
    split($1, parts, ":")
    filepath = parts[1]
    gsub(/github.com\/Notifuse\/liquidgo\//, "", filepath)
    
    # Extract coverage percentage
    coverage_str = $NF
    gsub(/%/, "", coverage_str)
    coverage = coverage_str + 0
    
    # If new file, print previous and reset
    if (filepath != file && file != "") {
        if (total > 0) {
            avg = (covered / total) * 100
            printf "%-50s %6.1f%%\n", file, avg
        }
        total = 0
        covered = 0
    }
    
    file = filepath
    total++
    covered += coverage / 100
}

END {
    if (total > 0 && file != "") {
        avg = (covered / total) * 100
        printf "%-50s %6.1f%%\n", file, avg
    }
}
