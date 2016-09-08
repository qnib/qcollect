#!/usr/bin/env gnuplot
set term png
set xtics rotate by 90 left
set style histogram gap 1
unset key
set yrange [0:100]
plot "coverity/cover_cur.out" using 2: xtic(1) with histogram, '' using 0:2:2 with labels
