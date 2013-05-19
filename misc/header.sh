# check for arguments
if [ $# -ne 1 ]
then
    echo "Error in $0 - Invalid Argument Count"
    echo "Syntax: $0 input_file - copyright header file"
    exit
fi

# check for header file
infile=$1
if [ ! -f $infile ]
then
    echo "Input file [$infile] not found - Aborting"
    exit
fi

# inject header
for i in *.go
do
  if ! grep -q Copyright $i
  then
    cat $infile $i >$i.new && mv $i.new $i
  fi
done
