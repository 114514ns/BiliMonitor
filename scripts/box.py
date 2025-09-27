import json

import pymysql

def query():
    cursor.execute("SELECT extra,gift_amount,gift_price FROM live_actions where action_type = 2 and extra like '%盲盒%' order by gift_price desc")
    array = cursor.fetchall()
    m = {}
    got = 0
    spent = 0
    count = 0
    t = {}
    watcher = {}
    for item in array:
        count = count + 1
        extra = item[0]
        name = str.split(extra,",")[0]
        got = got + int(item[2])
        if str.split(extra,",")[0] in t:
            t[name]["count"] = t[name]["count"] + 1
            t[name]["diff"] = t[name]["diff"] + int(item[2])-int(int(str.split(extra,",")[1]) * item[1])
        else:
            t[name] = {}
            t[name]["count"] = 1
            t[name]["returnRate"] = int(str.split(extra,",")[1])
            t[name]["diff"] = int(item[2])-int(int(str.split(extra,",")[1]) * item[1])
        spent = spent +int(int(str.split(extra,",")[1]) * item[1])

    for k in t:
        t[k]["percent"] = round(t[k]["count"]/count,2)
        t[k]["returnRate"] = round((t[k]["diff"] + t[k]["returnRate"] * t[k]["count"] ) / (t[k]["returnRate"] * t[k]["count"]),3)
    print(json.dumps(t, sort_keys=True, indent=4, separators=(',', ':'), ensure_ascii=False))
    print(f"unbox count {len(array)}")
    print(f"spend={spent}")
    print(f"got={got}")
    print(f"diff {got-spent}")
    print(f"return rate {got/spent}")




        
try:
    # 创建数据库连接
    connection = pymysql.connect(
        host="localhost",
        user="bili",
        password="CERtxrAH2zaik7b2",
        database="bili"
    )


    print("成功连接到数据库")

    # 创建一个游标对象
    cursor = connection.cursor()

    query()



except pymysql.MySQLError as e:
    print(f"数据库错误: {e}")
finally:
    # 关闭游标和连接
    if 'cursor' in locals() and cursor:
        cursor.close()
    if 'connection' in locals() and connection:
        connection.close()
        print("数据库连接已关闭")


if __name__ == '__main__':
    print()