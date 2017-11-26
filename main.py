import io
import subprocess
import json
import logging
import time
import traceback
from collections import defaultdict
from operator import itemgetter

import ijson.backends.yajl2_cffi as ijson
import requests
import schedule

_ESI_REGION_IDS_URL = "https://esi.tech.ccp.is/v1/universe/regions/"
_REGION_MARKET_URL_PATTERN = "https://element-43.com/api/orders/v2/region/{region_id}/"

def list_entry():
    return {
        'total_orders': 0,
        'total_volume': 0,
        'ask_volume': 0,
        'bid_volume': 0
    }

def update_stats():
    """
    Gets all regions from ESI, then gets orders from order service and aggregates statistics
    then saves them to a JSON file for retrieval vie Caddy.
    :return: None
    """

    logging.info("Updating stats...")

    try:
        # StationID -> Stats mapping, will be dumped to JSON as list later on
        top_stations = defaultdict(list_entry)

        # Get All regionIDs
        region_ids = requests.get(_ESI_REGION_IDS_URL, timeout=10).json()

        # Then get all region's markets and process orders
        for id in region_ids:
            logging.info("Getting %d..." % id)
            order_response = requests.get(_REGION_MARKET_URL_PATTERN.format(region_id=id), timeout=10)

            orders = ijson.items(io.BytesIO(order_response.content), 'item')

            for order in orders:
                stat_entry = top_stations[order['location_id']]
                stat_entry['total_orders'] += 1
                order_value = float(order['volume_remain'] * order['price'])

                stat_entry['total_volume'] += order_value

                if order['is_buy_order']:
                    stat_entry['bid_volume'] += order_value
                else:
                    stat_entry['ask_volume'] += order_value

        # Convert dict to list, sort by value and dump to file
        logging.debug("Saving stats..." )
        toplist = []
        for station_id, stats in top_stations.items():
            stats['location_id'] = station_id
            toplist.append(stats)

        toplist.sort(key=itemgetter('total_orders'), reverse=True)

        with open('stats/v2/list.json', 'w') as json_file:
            json.dump(toplist, json_file)

        logging.info("Done!")

    except:
        # I know this is bad, just don't crash
        traceback.print_exc()


logging.basicConfig(level=logging.INFO, format='time="%(asctime)s" msg="%(message)s"')

logging.debug("Spawning Caddy server...")
subprocess.Popen(["caddy"])
logging.debug("OK!")

update_stats()

schedule.every(60).minutes.do(update_stats)
while True:
    schedule.run_pending()
    time.sleep(1)
