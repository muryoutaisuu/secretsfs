cd $GOPATH/src/github.com/muryoutaisuu/secretsfs/
rm pdf/*
let c=1; for i in $(find . -name "*.go") ; do a2ps -o pdf/$c.ps $i; let c=$c+1; done;
cd pdf
let C=$(ls -l *ps | wc -l)
STRING="for i in {1..${C}}.ps; do ps2pdf \$i; done"
eval $STRING
#for i in {0..$(ls -l *.pdf | wc -l)}.pdf; do ps2pdf $i; done
cd ..
gs -q -sPAPERSIZE=a4 -dNOPAUSE -dBATCH -sDEVICE=pdfwrite -sOutputFile=sourcecode.pdf pdf/*.pdf
rm pdf/*
mv sourcecode.pdf pdf
