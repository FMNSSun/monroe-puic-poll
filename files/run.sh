#!/bin/sh

CERTS="/opt/monroe/rootCACert.pem"
COLLECT="256"
IFACE=$1
LOGFILE=""
ODIR="/var/monroe/tmp"
URLS=$2
RUNS="4"
RESULTDIR="/monroe/results"
WAITTO=1000
WAITFROM=2000

mkdir -p $ODIR

echo "ODIR: $ODIR"
echo "URLS: $URLS"

ls -lah /monroe/results
sleep 10s

while true
do
	echo "Invoking puic-poll..."
	/opt/monroe/puic-poll -certs=$CERTS -wait-to=$WAITTO -wait-from=$WAITFROM -collect=$COLLECT -iface=$IFACE -logfile=$LOGFILE -odir=$ODIR -urls=$URLS -runs=$RUNS
	ls -lah $ODIR
	echo "Moving results..."
	mv $ODIR/puic-poll* $RESULTDIR
	ls -lah $ODIR
	echo "Done..."
	sleep 5s
done
