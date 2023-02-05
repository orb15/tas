# TAS - Traveller Aid Software

This software presents a collection of command line tools to simplify working with the Mongoose Traveller 2E Update 2022 Roleplaying game.

# Command Directory
The sections below detail the various commands and their options.
You can always get help by executing `> tas -h` or `tas <command> -h` for help on a specific command.

### Gloabl Flags
These flags are available to every command  
`--loglevel <debug|info|warn|error|fatal>` flag can be used to set log level.
The default logging level is 'warn'  
`--tofile` when set, writes the output to a local output folder.
Output is in JSON, but is indented to make it easy to read

## world
The `world` command generates details of one or more worlds as expressed on pages 246 - 261 of the core rulebook.
Output is either a standard Universal World Profile (UWP - see pg 248) or a full-text display of the meaning behind each code.
An option is provided to use a custom world generation routine that generates more sensible world statistics (see world-debug command for more on this topic).

Usage: `> tas world [count] [flags]` where  

&nbsp;&nbsp;&nbsp;&nbsp;count is the number of worlds to generate. The default is 1.  
&nbsp;&nbsp;&nbsp;&nbsp;Flags:  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;`--long`  generate longform output instead of UWP.
Omitting this flag produces only UWP output  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;`--scheme <standard|custom>`
If this flag is included, one of the two options must be provided.
The default is 'standard' and uses the rules as written to generate worlds.
The 'custom' option utilizes a slightly different algorithm to generate more believable worlds.

## world debug (world sub-command)
The `world debug` sub-command isn't directly useful to sector designers, but instead is used to display the average stats of 40 (optionally: 10,000) randomly generated worlds.
This command exists to show averages and other stats for the randomly generated worlds in order to provide some insight into how well the rules-as-written generate random, believable, useful worlds.
Manual testing during development revealed that the worlds generated using the tables in the book had certain undesirable traits, including worlds with no atmosphere but meaningful water content and worlds with high populations or extreme environments yet very low tech levels.
The tech level algorithm is especially disturbing, as the rules as written generate an average tech level of 5 (mid 20th century tech), hardly sufficent to explain the hundreds of thousands of people living on an otherwise lifeless rock with no atmosphere.
This issue prompted the development of the "scheme" flag on the `world` command, allowing the use of customized generators to address these problems.
Addressing the hydrographics issue was easy, but I took a very hands-on approach to address the tech issue.
I set the base tech level to 7 (Pre-stellar, early 21st century Earth) as this was likely to be the lowest level of tech to be found on any world that wasn't subject to war or some other devestation.
From there, I used various traits about the world to drive the tech level up to allow life to exist, with each such increase directly tied to a higher level tech required to address problems presented by overpopulation, atmospherics, temperature or hydrographics.
For example, this means that if the world population is high on a desert world, the world will have a sufficient tech level to explain this apparent dichotomy.
The `world debug` sub-command is very useful for examining how a proposed algorithmic change to world generation actually impacts the kinds of worlds being generated.

Usage: `> tas world debug [flags]` where  

&nbsp;&nbsp;&nbsp;&nbsp;Flags:  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;`--scheme <standard|custom>`
If this flag is included, one of the two options must be provided.
The default is 'standard' and uses the rules as written to generate worlds.
The 'custom' option utilizes a slightly different algorithm to generate more believable worlds.
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;`--max`
If this flag is included, 10,000 worlds are used to generate stats rather than 40 (the averge number of worlds in a typical subsector). The differences between these are usually slight

Note: this command ignores the global `--tofile` flag!

## trade
The `trade` command calculates the Die Modifiers for the standard types of trade, including Passengers, Freight and Mail.
The entire process is not modelled (determining price, number of passengers etc.) because these depend on character skills and rolls.
Thus, the goal is more to do _most_ of the heavy lifting of calculating the final modifier while still allowing the characters to roll their skills.

Note that this command requires the use of a local data file stored in the 'data-local' folder.
This JSON file describes certain (typically) unchanging data about the universe where trade is conducted, including character skill data that impacts trade (e.g. Level of Steward, whether the ship is armed) and each world's UWP as a world's Population, Tech Level and Trade Codes affect trade in various ways.
It is from this JSON file that world data is extracted based on world name in the various `trade` commands.
An example file of this sort is given (see: data-local/example-trade-data.json); model any new files on the structure of the data in this file.
Also note that the UWP data in this example file is not meant to be correct / possible within the world generation system; it is just example data.

Best practice is to leave the example file intact and unedited and create your own file named 'trade-data.json' that mimics this file but contains all relevant real-game information.
If you do this, you do not need to set the `--file` flag (see Usage below), and the trade generation algorithm will use your data instead.

Usage: `> tas trade <current-world> <destination-world> [flags]` where

&nbsp;&nbsp;&nbsp;&nbsp;current-world is required and is the name of the world the player's are currently on  
&nbsp;&nbsp;&nbsp;&nbsp;destination-world is required and is the name of the world the player's are travelling to   
&nbsp;&nbsp;&nbsp;&nbsp;Flags:  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;`--file <filename>`
If this flag is set, the trade generator will use the given filename (rather than the default file 'trade-data.json') as a source of basic trade data

## trade spec (trade sub-command)
The `trade spec` sub-command generates the quantiity and purchase/sale DM's of the various trade Goods using the process outlined on pgs 241 - 245.
The algorithm calculates all available lots, including Common Goods, and Advanced or Illegal Goods that align to this world's Trade Codes and the random Trade Goods that just happen to be available on the current world.
Like the `trade` command, only DMs and other things not directly depending on character skill are calculated.
The algorithm saves the players and referee from having to exhaustively examine and re-examine the Trade Goods tables on pgs 244 - 245 to compare world Trade Codes against the Trade Codes where a given Trade Good might be found.
This sub-command uses the same data input file ("trade-data.json") as used for the main `trade` command. See that command for the purpose, name  and use of this file.

Note that this sub-command requires the use of the 'buy' or 'sell' argument. When the players are buying, a full list of Trade Lots is presented, along with Base Price, Quantity available and the DM to be used when determining the final Purchase Price.
The Sell command presents DMs for each possible Trade Good as defined on pgs 244 - 245.
Players can use these DM's to determine the Offered Purchase Price some NPC agent is willing to pay for the goods they own.
In both cases, the DM's provided are not the final DM's; player skill level, the use of a local broker (or underworld fixer in the case of  Illegal Goods) or in-universe reasons may adjust this DM before it is used to determine price information.

Note that while Illegal Goods are likely to show up as an available Trade Lot, they are not actually made available to the players without roleplay or use of a local 'fixer' or underworld broker to provide access to these goods. Finally, note that some goods that are generally considered legal on a generic world (e.g. Weapons) may be illegal on a given world based on Law Level or other in-universe reasons.
The referee will need to determine if the goods are just not available at all, or if they are available but heavily controlled to ensure their immediate export off world.

Usage: `> tas trade spec <current-world> <buy|sell> [flags]` where  
&nbsp;&nbsp;&nbsp;&nbsp;current-world is required and is the name of the world the player's are currently on  
&nbsp;&nbsp;&nbsp;&nbsp;buy or sell is required and indicates whether the players are looking to BUY goods on the current world or SELL goods they already own on the current world  
&nbsp;&nbsp;&nbsp;&nbsp;Flags:    
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;`--file <filename>`
If this flag is set, the trade generator will use the given filename (rather than the default file 'trade-data.json') as a source of basic trade data