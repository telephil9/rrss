     RRSS(1)                                                       RRSS(1)

     NAME
          rrss, trrss - RSS feed readers

     SYNOPSIS
          rrss [-f barf|blagh] [-r root] [-t tag] [-u url]

          trrss [-f barf|blagh] [-r root] [-t tag] [-u url]

     DESCRIPTION
          Rrss pulls and parses an RSS feed.

          There are a number of options:

          -f   Place output in formatted directories for one
               of two werc apps: barf or blagh. In the absence
               of the -f flag, formatted output is placed on
               stdout.

               A file, links, is created in the root and is populated
               with the URL of each feed item acquired. On sub-
               sequent runs, URLs that appear in the links file are
               not duplicated as new directories.

          -r   Optionally, create barf or blagh directories
               under root. Default is the current directory.

          -t   Create tag for each post (barf only).

          -u   The feed URL.

          Trrss is a shell script that wraps the rrss program,
          outputting plain text but preserving link URLs.

     SOURCE
          http://plan9.stanleylieber.com/src/rrss.tgz
     SEE ALSO
          http://werc.cat-v.org
          http://werc.cat-v.org/apps/blagh
          https://code.9front.org/hg/barf
