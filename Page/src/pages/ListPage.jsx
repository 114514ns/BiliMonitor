import React, {memo, useEffect, useLayoutEffect} from 'react';
import {
    Autocomplete,
    AutocompleteItem,
    Avatar,
    Card,
    CardBody,
    Chip, Input,
    Listbox,
    ListboxItem, Select, SelectItem,
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
    
    const [filted, setFiltered] = React.useState([]);

    const [verify,setVerify] = React.useState([]);

    const [verifyFilter, setVerifyFilter] = React.useState('');
    const [nameFilter, setNameFilter] = React.useState('');
    const [bioFilter,setBioFilter] = React.useState('');


    const port = location.port

    const protocol = location.protocol.replace(":", "")
    useEffect(() => {
        var url = `${protocol}://${host}:${port}/api/areaLivers`
        axios.get(url).then((response) => {
            setList(response.data.list);
            setFiltered(response.data.list);
            var map = new Map();

            response.data.list.forEach(item => {
                item.Verify.split("、").forEach(e => {
                    if (e!=="") {
                        if (map.has(e)) {
                            map.set(e, map.get(e)+1);
                        } else {
                            map.set(e,1)
                        }
                    }

                })
            })
            var temp = []
            map.forEach((item, i) => {
                temp.push(i);
            })
            var array =Array.from(map);
            temp = ['Any']
            array.sort((a,b)=>{
                return b[1]-a[1];
            }).forEach(e => {
                temp.push(e[0]);
            })
            setVerify(temp)
        })
    },[])

    useEffect(() => {
        var o = list
        if (nameFilter != '') {
            o = o.filter(i => { return i.UName.indexOf(nameFilter) !== -1 })
        }

        if(verifyFilter!==''&&verifyFilter!=='Any'){
            o = o.filter(i => { return i.Verify.indexOf(verifyFilter) !== -1 })
        }
        if (bioFilter != '') {
            o = o.filter(i => { return i.Bio.indexOf(bioFilter) !== -1 })
        }
        setFiltered(o)
    },[verifyFilter,nameFilter,bioFilter])
    return (
        <div>
            <Autocomplete
                className="max-w-xs"
                defaultItems={sort}
                label="Sort by"
                placeholder="粉丝"
                style={{
                    marginLeft:'4px'
                }}
            >
                {(sort) => <AutocompleteItem key={sort.key} onPress={(e) => {
                    var url = `${protocol}://${host}:${port}/api/areaLivers?sort=${sort.key}`
                    axios.get(url).then((response) => {
                        setList(response.data.list);
                        setFiltered(response.data.list);
                    })
                    console.log(sort.key);
                }}>{sort.description}</AutocompleteItem>}
            </Autocomplete>
            <Select className="max-w-xs" label="Verify filter" placeholder="">
                {verify.map((item) => (
                    <SelectItem key={item} onPress={e => setVerifyFilter(e.target.innerText)}>{item}</SelectItem>
                ))}
            </Select>
            <Input className='max-w-xs' onChange={event => setBioFilter(event.target.value)}></Input>
            <Listbox
                virtualization={{
                    maxListboxHeight: window.innerHeight,
                    itemHeight: 300,
                }}
                hideSelectedIcon
                variant={'light'}
                isVirtualized>
                {filted.map((item, index) => (
                    <ListboxItem key={index} value={item.value} css={{width:'100%'}} aria-label={item.label} textValue={''}>
                        <LiverCard
                            Rank={index}
                            Avatar={`${protocol}://${host}:${port}${import.meta.env.PROD?'':'/api'}/face?mid=${item.UID}`}
                            UName={item.UName}
                            Guard={item.Guard}
                            DailyDiff={item.DailyDiff}
                            Fans={item.Fans}
                            LastActive={(item.LastActive)}
                            UID={item.UID}
                            Bio={item.Bio}
                            Verify={item.Verify}

                        />
                    </ListboxItem>))}
            </Listbox>

            <div style={{
                position: 'fixed',
                right:'20px',
                bottom:'20px',
                width:'180px',
                height:'60px',
            }}>
                <Input label="Search"  onValueChange={(e) => {
                    setNameFilter(e)

                }}/>
            </div>
        </div>
    );

}

const LiverCard = memo(function LiverCard(props) {
    const up = props.DailyDiff>=0
    return (
        <Card isHoverable  style={{ width: "100%", marginTop: "16px" }} >
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


                <div style={{ display: "flex", flexDirection: "column", alignItems: "center", lineHeight: 1.6 }}>
                    <p style={{ margin: 0 }}>关注：{formatNumber(props.Fans)}</p>
                    <p style={{ margin: 0 }}>日增：{<span style={{color:up?'#00cc00':'#ff0000'}}>{up?'▲':'▼'}</span>}{props.DailyDiff}</p>
                    <p style={{ margin: 0 }}>大航海：{props.Guard}</p>
                    <p style={{ margin: 0, color: "#888" }}>上次直播：{formatTime(props.LastActive)}</p>
                    <p>{props.Bio}</p>
                    {props.Verify===''?<></>:<p style={{color:'rgba(190,151,48,1)'}}>{props.Verify}</p>}
                    <p>Rank:{props.Rank+1}</p>
                </div>

            </CardBody>
        </Card>
    )
})
export default ListPage;