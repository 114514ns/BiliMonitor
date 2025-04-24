import React, {useEffect, useLayoutEffect} from 'react';
import {
    Autocomplete,
    AutocompleteItem,
    Avatar,
    Card,
    CardBody,
    Chip,
    Listbox,
    ListboxItem,
    Spacer
} from "@heroui/react";
import axios from "axios";
function formatTime(isoString) {
    const date = new Date(isoString);

    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day   = String(date.getDate()).padStart(2, '0');
    const hour  = String(date.getHours()).padStart(2, '0');
    const min   = String(date.getMinutes()).padStart(2, '0');

    return `${month}月${day}日 ${hour}:${min}`;
}
function formatNumber(num) {
    if (num >= 10000) {
        return (num / 10000).toFixed(1).replace(/\.0$/, '') + '万';
    } else {
        return String(num);
    }
}
const sort = [
    {label: "guard", key: "guard", description: "大航海"},
    {label: "l1-guard", key: "l1-guard", description: "总督"},
    {label: "fans", key: "fans", description: "粉丝"},
    {label: "diff", key: "diff", description: "日增"},
    {label: "guard-equal", key: "guard-equal", description: "等效舰长"},
];
function ListPage(props) {
    const [list, setList] = React.useState([]);
    const host = location.hostname;


    const port = location.port

    const protocol = location.protocol.replace(":", "")
    useEffect(() => {
        var url = `${protocol}://${host}:${port}/api/areaLivers`
        axios.get(url).then((response) => {
            setList(response.data.list);
        })
    },[])
    return (
        <div>
            <Autocomplete
                className="max-w-xs"
                defaultItems={sort}
                label="排序方式"
                placeholder="粉丝"
                style={{
                    marginLeft:'4px'
                }}
            >
                {(sort) => <AutocompleteItem key={sort.key} onPress={(e) => {
                    var url = `${protocol}://${host}:${port}/api/areaLivers?sort=${sort.key}`
                    axios.get(url).then((response) => {
                        setList(response.data.list);
                    })
                    console.log(sort.key);
                }}>{sort.description}</AutocompleteItem>}
            </Autocomplete>
            <Listbox
                virtualization={{
                    maxListboxHeight: window.innerHeight,
                    itemHeight: 240,
                }}
                hideSelectedIcon
                variant={'light'}
                isVirtualized>
                {list.map((item, index) => (
                    <ListboxItem key={index} value={item.value} css={{width:'100%'}} aria-label={item.label}>
                        <LiverCard
                            Avatar={`${protocol}://${host}:${port}/face?mid=${item.UID}`}
                            UName={item.UName}
                            Guard={item.Guard}
                            DailyDiff={item.DailyDiff}
                            Fans={item.Fans}
                            LastActive={(item.LastActive)}
                            UID={item.UID}

                        />
                    </ListboxItem>))}
            </Listbox>
        </div>
    );

}

function LiverCard(props) {
    return (
        <Card isHoverable  style={{ width: "100%", marginTop: "16px" }}  isPressable>
            <CardBody style={{ display: "flex", alignItems: "center", justifyContent: "space-between", width: "100%", padding: "16px" }}>


                <div style={{ display: "flex", alignItems: "center" }}>
                    <Avatar
                        src={props.Avatar}
                        alt={props.UName}
                        width={64}
                        height={64}
                        radius="full"
                    />
                    <div style={{ marginLeft: "12px" }}>
                        <p style={{ fontSize: "16px", fontWeight: "500", margin: 0 }}>{props.UName}</p>
                    </div>
                </div>
                <Chip style={{margin:'4px'}} onClick={() => {toSpace(props.UID)}}>{props.UID}</Chip>


                <div style={{ display: "flex", flexDirection: "column", alignItems: "flex-end", lineHeight: 1.6 }}>
                    <p style={{ margin: 0 }}>关注：{formatNumber(props.Fans)}</p>
                    <p style={{ margin: 0 }}>日增：{props.DailyDiff}</p>
                    <p style={{ margin: 0 }}>大航海：{props.Guard}</p>
                    <p style={{ margin: 0, color: "#888" }}>上次直播：{formatTime(props.LastActive)}</p>
                </div>

            </CardBody>
        </Card>
    )
}
export default ListPage;