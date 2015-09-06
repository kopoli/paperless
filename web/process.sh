#!/bin/sh

# Processes a single image to the following:
# - cleaned up image
# - ocr'd text
# - thumbnail

# configuration
def() { eval "$1=\${$1-\"$2\"}"; }

def TMP_PATTERN /tmp/img.XXXXXXXX
def UNPAPER unpaper
def TESSERACT tesseract
def CONVERT convert

def TESSERACT_ARGS "-l fin -psm 1"
def UNPAPER_ARGS "-vv -s a4 -l single -dv 3.0 -dr 80.0 --overwrite"

def CONVERT_FMT_UNPAPER pnm
def CONVERT_FMT_TESSERACT pnm
def CONVERT_FMT_FINAL jpg
def TESSERACT_FMT txt

def CONVERT_ARGS_UNPAPER "-depth 8"
def CONVERT_ARGS_TESSERACT "-normalize -colorspace Gray"
def CONVERT_ARGS_FINAL "-trim -quality 80% +repage -type optimize"

def CONVERT_ARGS_THUMB "$CONVERT_ARGS_FINAL -thumbnail 200x200>"

def THUMB_IMG_ID "-thumb"

# Logs the output of a command
# args: args ...
log()
{
    echo "$@"
    ( "$@" ) 2>&1 || fail=1
    if test -n "$fail"; then
        echo "Failed command: $@"
        return 1
    fi
    return 0
}

process_single_deinit()
{
    ret=$?
    rm -f $TMP_UNPAPER_IMG $TMP_CONVERT_IMG $TMP_TESSERACT_IMG $TGTTXT ${TGTTXT%.*}${TESSERACT_FMT}
    exit $ret
}

# main script

FILE="$1"
TGTIMG="$2"
THUMB="$3"
# def TGTIMG "${BASE}.${CONVERT_FMT_FINAL}"
# def THUMB "${BASE}${THUMB_IMG_ID}.${CONVERT_FMT_FINAL}"

do_ocr=t
do_convert=t

test -z "$FILE" && exit 1

trap process_single_deinit INT TERM EXIT

echo "Processing $FILE"

set -e
TMP_UNPAPER_IMG=$(mktemp $TMP_PATTERN.${CONVERT_FMT_UNPAPER})
log $CONVERT $CONVERT_ARGS_UNPAPER $FILE $TMP_UNPAPER_IMG

TMP_CONVERT_IMG=$(mktemp $TMP_PATTERN.${CONVERT_FMT_TESSERACT})
log $UNPAPER $UNPAPER_ARGS $TMP_UNPAPER_IMG $TMP_CONVERT_IMG

TGTTXT=$(mktemp $TMP_PATTERN)

TMP_TESSERACT_IMG=$(mktemp $TMP_PATTERN.${CONVERT_FMT_TESSERACT})
if test -n "$do_ocr"; then
    log $CONVERT $CONVERT_ARGS_TESSERACT $TMP_CONVERT_IMG $TMP_TESSERACT_IMG
    log $TESSERACT -l fin $TMP_TESSERACT_IMG ${TGTTXT%.*} $TESSERACT_ARGS $TESSCONF
fi
if test -n "$do_convert"; then
    log $CONVERT $CONVERT_ARGS_FINAL $TMP_CONVERT_IMG $TGTIMG
    log $CONVERT $CONVERT_ARGS_THUMB $TMP_CONVERT_IMG $THUMB
fi

echo "__LOG_ENDS_HERE__"

cat ${TGTTXT%.*}.${TESSERACT_FMT}

set +e
