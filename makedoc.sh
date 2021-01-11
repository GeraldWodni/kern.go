#!/bin/bash
# glue around godocdown to get a single markdown file for kern.go
# required: godocdown, marked

# target filename
MARKDOWN=ReadMe.md
HTML=ReadMe.html

function doc {
    # remove "--" and copyrights from documentation
    godocdown $* \
        | grep -v 'import "\.\"' \
        | grep -v '^--$' \
        | sed -e 's/(c).*>//'
}

function separator {
echo -e "\n---\n" >> $MARKDOWN
}

# add module here
doc . > $MARKDOWN
separator
doc ./router >> $MARKDOWN
separator
doc ./view >> $MARKDOWN
separator
doc ./session >> $MARKDOWN
separator
doc ./login >> $MARKDOWN
separator
doc ./log >> $MARKDOWN

cat $MARKDOWN | marked --gfm > $HTML
