WARNING: Not yet ready for use, or even just functional

# Lamprey

A wiki software that allows for easy hosting of static pages and dynamic content. It combines text
based articles with JSON based machine readable data.

It's written in go and aims to be as trivial to host as possible.

## Entries

The wiki consists of entries. Every entry has two parts: An Article, containing (partially
generated) HTML content, and Data.

## Usecases I'd like to see covered

- Database like websites. Should comfortably represent data of Yu-Gi-Oh wiki, Minecraft wiki,
  Pokewiki
- Game maps (A generated SVG on a base bitmap, with specific points or polygons being links and/or
  having tooltips)

# Articles

- Mostly normal text. Can contain templates that automatically generate markup from the entries
  Data.
- Each entry has multiple articles, one for each supported language
- They are statically generated and don't need any dynamic server code to be read

# Data

- The data of an entry is kept in JSON.
- They consist of several Data Blocks, each of which follows a predefined schema.
- Entry Data can contain multiple Data Blocks of the same schema
  - (Should the schema be able to pose any limitation on that? Like a unique key?)
    - (Maybe even fully support DB like linking? Like, one data block in one Entry automatically
      creates a corresponding Data Block in a linked entry)
- Example:
  - Yu-Gi-Oh cards have multiple names in different languages. They can also change names over
    time. So they could have a "name" scheme, with a language, the name and an optional date range
  - The same animals appear across multiple entries in the Animal Crossing series, with different
    museum descriptions in different versions and localizations. So they could have a "museum text"
    scheme, with a language, a game version and the text.
  - A game map could use multiple entries of a "location" schema, containing coordinates on a map
    and a link to another entry.

# Hosting and performance considerations

- When pages are saved, they are generated as static pages. Read access to the wiki doesn't utilize
  the wikisoftware at all.
- JS is optional, but allows cleaner integration of static pages into the dynamic system
- Data is similarly saved as static JSON
- For the start, no Data Query Interface will be provided. But there should be a daily dump
  feature. From there, the files can e.g. be put onto bittorrent to allow for convenient download.
- The software should support different languages per wiki, and different wikis on one server.
- The dynamic section can probably use something like CGI to integrate. Alternatively, it should be
  able to live on a completely separate endpoint, so you can host a Read-Only wiki with intranet
  write access.

# Mission statement

- The target isn't MediaWiki/Wikibase/Wikipedia/Wikidata
- The target is the ad infested hell site Fandom
- To be able to beat the service, hosting must be as trivial and cheap as possible. Even with
  proper caching, I feel like something based on MediaWiki shouldn't be able to beat static hosting
- Many wikis about games, shows, etc. also collect a lot of factual, structured data, that's
  trapped in difficult to parse markup.
- Wikibase exists, but seems to complicated for many to make use of it.
