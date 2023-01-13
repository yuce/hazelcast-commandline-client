package python

const sampleNotebookName = "MyNotebook"

const sampleNotebook = `
{
 "cells": [
  {
   "cell_type": "markdown",
   "id": "7c20eec1",
   "metadata": {},
   "source": [
    "# My Hazelcast Notebook\n",
    "\n",
    "![Hazelcast Logo](https://hazelcast.com/wp-content/themes/hazelcast/img/icons/favicons/apple-touch-icon.png)\n",
    "\n",
    "Welcome to your Hazelcast Notebook.\n",
    "\n",
    "Using this Notebook, you can do:\n",
    "\n",
    "* [Real-Time Streaming Analytics](https://hazelcast.com/use-cases/real-time-streaming-analytics/)\n",
    "* [Real-Time Stream Processing](https://hazelcast.com/use-cases/real-time-stream-processing/)\n",
    "* [Fast Batch Processing](https://hazelcast.com/use-cases/fast-batch-processing/)\n",
    "\n",
    "Check out our [Python documentation](https://hazelcast.readthedocs.io/en/stable/).\n",
    "\n",
    "Here are a few cells to get you started."
   ]
  },
  {
   "cell_type": "code",
   "execution_count": null,
   "id": "ed199b0f",
   "metadata": {},
   "outputs": [],
   "source": [
    "# get a connection to the Hazelcast cluster\n",
    "from clc import conn"
   ]
  },
  {
   "cell_type": "code",
   "execution_count": null,
   "id": "ca97358f",
   "metadata": {},
   "outputs": [],
   "source": [
    "# get a cursor\n",
    "cursor = conn.cursor()"
   ]
  },
  {
   "cell_type": "code",
   "execution_count": null,
   "id": "6ce4a144",
   "metadata": {},
   "outputs": [],
   "source": [
    "# create a mapping\n",
    "cursor.execute('''\n",
    "    CREATE OR REPLACE MAPPING city (\n",
    "    __key INT,\n",
    "    country VARCHAR,\n",
    "    name VARCHAR)\n",
    "    TYPE IMap\n",
    "    OPTIONS('keyFormat'='int', 'valueFormat'='json-flat');\n",
    "''')"
   ]
  },
  {
   "cell_type": "code",
   "execution_count": null,
   "id": "1a23f7c2",
   "metadata": {},
   "outputs": [],
   "source": [
    "# add some data\n",
    "cursor.execute('''\n",
    "    SINK INTO city VALUES\n",
    "    (1, 'United Kingdom','London'),\n",
    "    (2, 'United Kingdom','Manchester'),\n",
    "    (3, 'United States', 'New York'),\n",
    "    (4, 'United States', 'Los Angeles'),\n",
    "    (5, 'Turkey', 'Ankara'),\n",
    "    (6, 'Turkey', 'Istanbul'),\n",
    "    (7, 'Brazil', 'Sao Paulo'),\n",
    "    (8, 'Brazil', 'Rio de Janeiro');\n",
    "''')"
   ]
  },
  {
   "cell_type": "code",
   "execution_count": null,
   "id": "20a66195",
   "metadata": {},
   "outputs": [],
   "source": [
    "# run a query\n",
    "cursor.execute('''\n",
    "    SELECT * FROM city;\n",
    "''')\n",
    "for row in cursor:\n",
    "    print(row[\"__key\"], row[\"name\"], row[\"country\"], sep=\"\\t\")"
   ]
  },
  {
   "cell_type": "code",
   "execution_count": null,
   "id": "d5c7003d",
   "metadata": {},
   "outputs": [],
   "source": [
    "# create a data frame from SQL\n",
    "import pandas\n",
    "df = pandas.read_sql(\"select * from city\", conn)"
   ]
  },
  {
   "cell_type": "code",
   "execution_count": null,
   "id": "78a0fa48",
   "metadata": {},
   "outputs": [],
   "source": [
    "df"
   ]
  },
  {
   "cell_type": "code",
   "execution_count": null,
   "id": "5639da0a",
   "metadata": {},
   "outputs": [],
   "source": [
    "# create a plot from the dataframe\n",
    "df.plot()\n"
   ]
  }
 ],
 "metadata": {
  "kernelspec": {
   "display_name": "Python 3 (ipykernel)",
   "language": "python",
   "name": "python3"
  },
  "language_info": {
   "codemirror_mode": {
    "name": "ipython",
    "version": 3
   },
   "file_extension": ".py",
   "mimetype": "text/x-python",
   "name": "python",
   "nbconvert_exporter": "python",
   "pygments_lexer": "ipython3",
   "version": "3.10.6"
  }
 },
 "nbformat": 4,
 "nbformat_minor": 5
}
`
