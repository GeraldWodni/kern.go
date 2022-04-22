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

function toc  {
    echo -e -n "- [kern](#kern)\\\\n"
    for MODULE in $MODULES; do
        echo -e -n "- [$MODULE](#$MODULE)\\\\n"
    done
}

function separator {
echo -e "\n---\n" >> $MARKDOWN
}

# add module here
MODULES="
    router
    view
    hierarchy
    redis
    filter
    session
    module
    login
    logout
    log
    "

doc . > $MARKDOWN


# module documentation
for MODULE in $MODULES; do
    echo "Module: $MODULE"
    separator
    doc ./$MODULE >> $MARKDOWN
done

# toc
TOC=$(toc)
sed -i -e "s/# Documentation/# Documentation\\n $TOC/" $MARKDOWN

cat $MARKDOWN | marked --gfm > $HTML
