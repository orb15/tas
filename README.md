# TAS - Traveller Aid Software

This software presents a collection of command line tools to simplify working with the Mongoose Traveller 2E Update 2022 Roleplaying game.

# Command Directory
The sections below detail the various commands and their options.
You can always get help by executing `> tas -h` or `tas <command> -h` for help on a specific command.

### Global Flags
These flags are available to every command  
`--loglevel <debug|info|warn|error|fatal>` flag can be used to set log level.
The default logging level is 'warn'  
`--tofile` when set, writes the output to a local output folder.
Output is in JSON, but is indented to make it easy to read

---
## world
The `world` command generates details of one or more worlds as expressed on pages 246 - 261 of the core rulebook.
Output is either a standard Universal World Profile (UWP - see pg 248) or a full-text display of the meaning behind each code.
Note that several option calculations: bases, temperature, factions, and culture are not calculated. These details existed in previous versions of this code, were determined to be not useful, and just cluttered the world longform profile with information that is probably better determined by the GM on a case-by-case basis.

Usage: `> tas world [count] [flags]` where  

&nbsp;&nbsp;&nbsp;&nbsp;count is the number of worlds to generate. The default is 1.  
&nbsp;&nbsp;&nbsp;&nbsp;Flags:  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;`--long`  generate longform output instead of UWP.
Omitting this flag produces only UWP output  

## world debug (world sub-command)
The `world debug` sub-command isn't directly useful to sector designers, but instead is used to display the average stats of 40 (optionally: 10,000) randomly generated worlds.
This command exists to show averages and other stats for the randomly generated worlds in order to provide some insight into how well the rules-as-written generates random, believable, useful worlds.
Manual testing during development revealed that the worlds generated using the tables in the book had certain undesirable traits, including worlds with no atmosphere but meaningful water content and worlds with high populations or extreme environments yet very low tech levels.
The tech level algorithm is especially disturbing, as the rules as written generate an average tech level of 5 (mid 20th century tech), hardly sufficent to explain the hundreds of thousands of people living on an otherwise lifeless rock with no atmosphere.

The `world debug` sub-command is very useful for examining how a proposed algorithmic change to world generation actually impacts the kinds of worlds being generated.

Usage: `> tas world debug [flags]` where  

&nbsp;&nbsp;&nbsp;&nbsp;Flags:  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;`--worldscheme <standard|custom>`
If this flag is included, one of the two options must be provided.
The default is 'standard' and uses the rules as written to generate worlds.
The 'custom' option utilizes a slightly different algorithm to generate more believable worlds.
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;`--max`
If this flag is included, 10,000 worlds are used to generate stats rather than 40 (the averge number of worlds in a typical subsector). The differences between these are usually slight

Note: this command ignores the global `--tofile` flag!

---

## sector
The `sector` command generates a subsector (8x10 hex grid), creating full world details, establishing the presence of gas giants and the like.
The approach taken is to use the algorithm for world density and gas giant presence as expressed on pg 246.
This is accomplished largely by generating a set of random worlds with names and sub-sector locations, which uses the algorithms expressed by the `world` command above.
Since this command uses the world generation described above, it uses several of the same flags as well as the global flags.

Usage: `> tas sector <sector-name> [flags]` where  

&nbsp;&nbsp;&nbsp;&nbsp;sector-name is required and is the name of this sector  
&nbsp;&nbsp;&nbsp;&nbsp;Flags:  

If the global `--tofile` flag is used, a folder with the sector's name is written into the output directory and all world related-data is written into individual files in this folder.

---

## polish
The `polish` utility command reads the ./data-local/world-names.txt file and performs the following clean-up on the file:

* Removes duplicate world names
* Trims excess spaces from world names
* Sorts world names alphabetically

It does this is a very safe manner (through the use of temporary files) so a huge list of world names is not accidently overwritten or discarded.

Usage: `> tas polish`  
No flags or arguments are accepted or honored.

Note that from time to time, this list of worlds will be extended (and perhaps trimmed).
Be careful when pulling the latest version of this code so as not to destroy your list of world names!