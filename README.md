# bitindex

[![Build Status](https://travis-ci.org/chop-dbhi/bitindex.svg?branch=master)](https://travis-ci.org/chop-dbhi/bitindex) [![GoDoc](https://godoc.org/github.com/chop-dbhi/bitindex?status.svg)](https://godoc.org/github.com/chop-dbhi/bitindex)

A small library for building and operating on bit indexes. A *bit index* of composed of a **domain**  and a **table**. A domain maps a set of `uint32` members to bit positions. The table contains a set of sparse bit arrays for each key in the index.

For example, given a set of fruit choices (called the domain):

- Apples
- Cherries
- Peaches
- Grapes

and a set of people (referred to as keys) with the fruit they enjoy (referred to at the item set):

- Bob: [Apples, Peaches]
- Sue: [Grapes]
- Joe: [Grapes, Cherries, Peaches]

The resulting matrix would look like this:

 |Apples|Cherries|Peaches|Grapes
----|------|--------|-------|------
Bob|1|0|1|0
Sue|0|0|0|1
Joe|0|1|1|1

Once the index is built, lookup operations can be performed to find the items that match certain criteria:

- **any** - Find all keys that have membership for any item in the lookup set.
- **all** - Find all keys that have membership for all items in the lookup set.
- **not any** - Find all keys that do not have membership for any item in the lookup set.
- **not all** - Find all keys that do not have membership for all items in the lookup set.

Applying these operations to the example:

- People who enjoy Apples or Cherries (an *any* operation) would match Bob and Joe.
- People who enjoy Apples and Peaches (an *all* operation) would match Bob.
- People who *do not* enjoy Peaches or Apples (a *not any* operation) would match Sue.
- People who *do not* enjoy Grapes and Cherries (a *not all* operation) would match Bob and Sue.

## Install

Download a release from the [releases page](https://github.com/chop-dbhi/bitindex/releases) for your architecture.

### Source

Clone the repository and run the following to install `bitindex` in your `$GOPATH/bin` directory.

```
make install build
```

## Usage

### Build an index

To build an index, the input stream must be in one of the supported [formats](#formats). For this example, we will use a CSV format and assuming the fruit and people have the following IDs.

ID|Fruit
----|-----
1|Apples
2|Cherries
3|Peaches
4|Grapes

ID|Person
----|------
100|Bob
101|Sue
102|Joe

An index requires use of unsigned 32-bit integers, so we can use the IDs to encoded the labels.

```sh
cat << EOF > fruit.csv
fruit,person
1,100
3,100
4,101
4,102
2,102
3,102
EOF
```

To build the index, use the `build` command. Since we included a CSV header, we add the flag to denote that. The index is written to stdout by default, but we included the `--output` option to specify a filename. To denote which column is the domain and which is the set key, we add the corresponding flags with the column index.

```
bitindex build --format=csv --csv-header --csv-domain=0 --csv-key=1 --output=fruit.bitx fruit.csv
```

This will output the following 

```
Build time: 22.684µs
Statistics
* Domain size: 4
* Table size: 3
* Sparsity: 0
```

The domain size is equal to the number of fruit and the table size is the number of people.

### Query the index

As noted above, the four operations that are supported are `any`, `all`, `nany`, and `nall`. Any or all of these flags can be passed with a comma-separated list of fruit IDs. Below query the index for the questions asked above.

Apples or Cherries

```
bitindex query --any=1,2
100
102
```

Apples and Peaches

```
bitindex query --all=1,3
100
```

Not (Peaches or Apples)

```
bitindex query --nany=3,1
101
```

Not (Grapes and Cherries)

```
bitindex query --nall=4,2
100
101
```

## Formats

Currently, the only supported format is CSV.
