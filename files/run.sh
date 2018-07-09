#!/bin/sh

ODIR="/var/monroe/tmp"
RESULTDIR="/monroe/results"
CERTS="" #"/opt/monroe/rootCACert.pem"

mkdir -p $ODIR

echo "ODIR: $ODIR"
echo "URLS: $URLS"

ls -lah /monroe/results

echo "Invoking puic-poll..."
/opt/monroe/puic-poll -odir=$ODIR -certs=$CERTS
ls -lah $ODIR
echo "Moving results..."
mv $ODIR/puic-poll* $RESULTDIR
ls -lah $ODIR
echo "Done..."

