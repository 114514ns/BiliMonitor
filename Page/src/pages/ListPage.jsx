import React, {memo, useEffect} from 'react';
import {
    addToast,
    Avatar,
    Button,
    Card,
    CardBody, CardHeader,
    Chip,
    Input,
    Listbox,
    ListboxItem,
    Select,
    SelectItem, ToastProvider, Tooltip,Code
} from "@heroui/react";
import axios from "axios";
import alasql from "alasql";
import {useNavigate} from "react-router-dom";

function formatTime(isoString) {
    const date = new Date(isoString);

    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hour = String(date.getHours()).padStart(2, '0');
    const min = String(date.getMinutes()).padStart(2, '0');

    return `${month}月${day}日 ${hour}:${min}`;
}

function formatNumber(num) {
    if (num >= 10000) {
        return (num / 10000).toFixed(1).replace(/\.0$/, '') + '万';
    } else {
        return String(num);
    }
}
const calcHeight = () => {
    const vh = window.innerHeight;
    const rem = parseFloat(getComputedStyle(document.documentElement).fontSize);
    const result = vh - 4 * rem;
    return result;
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

    const [verify, setVerify] = React.useState([]);

    const [verifyFilter, setVerifyFilter] = React.useState('');
    const [nameFilter, setNameFilter] = React.useState('');
    const [bioFilter, setBioFilter] = React.useState('');


    const port = location.port

    const protocol = location.protocol.replace(":", "")

    const redirect = useNavigate();


    var rawSQLRef = React.createRef();


    useEffect(() => {
        var url = `${protocol}://${host}:${port}/api/areaLivers`
        axios.get(url).then((response) => {
            response.data.list.forEach((element, index) => {
                const parts = element.Guard.split(',');
                response.data.list[index].GuardCount =
                    parseInt(parts[0]) + parseInt(parts[1]) + parseInt(parts[2]);
            });
            setList(response.data.list);
            setFiltered(response.data.list);
            var map = new Map();

            response.data.list.forEach(item => {
                item.Verify.split("、").forEach(e => {
                    if (e !== "") {
                        if (map.has(e)) {
                            map.set(e, map.get(e) + 1);
                        } else {
                            map.set(e, 1)
                        }
                    }

                })
            })
            var temp = []
            map.forEach((item, i) => {
                temp.push(i);
            })
            var array = Array.from(map);
            temp = ['Any']
            array.sort((a, b) => {
                return b[1] - a[1];
            }).forEach(e => {
                temp.push(e[0]);
            })
            setVerify(temp)
        })
    }, [])

    useEffect(() => {
        var o = list
        if (nameFilter != '') {
            o = o.filter(i => {
                return i.UName.indexOf(nameFilter) !== -1
            })
        }

        if (verifyFilter !== '' && verifyFilter !== 'Any') {
            o = o.filter(i => {
                return i.Verify.indexOf(verifyFilter) !== -1
            })
        }
        if (bioFilter != '') {
            o = o.filter(i => {
                return i.Bio.indexOf(bioFilter) !== -1
            })
        }
        setFiltered(o)
    }, [verifyFilter, nameFilter, bioFilter])

    var inputRef = React.createRef();


    return (

        <div>
            <div style={{display: "flex"}} className='flex-col sm:flex-row sm:align-items-center' ref={inputRef}>
                <Select
                    className="max-w-xs mb-4 mr-4"
                    label="Sort by"
                    placeholder="粉丝"
                    style={{
                        marginLeft: '4px'

                    }}
                >

                    {sort.map((item) => (
                        <SelectItem key={item.key} onPress={(e) => {
                            var url = `${protocol}://${host}:${port}/api/areaLivers?sort=${item.key}`
                            axios.get(url).then((response) => {
                                setList(response.data.list);
                                setFiltered(response.data.list);
                            })
                            console.log(item.key);
                        }}>{item.description}</SelectItem>
                    ))}
                </Select>
                <Select className="max-w-xs mb-4 mr-4" label="Verify filter" placeholder="">
                    {verify.map((item) => (
                        <SelectItem key={item} onPress={e => setVerifyFilter(e.target.innerText)}>{item}</SelectItem>
                    ))}
                </Select>
                <Input className='max-w-xs mb-4 mr-4' onChange={event => setBioFilter(event.target.value)}
                       label={'Sign filter'}></Input>
                <Tooltip content={<Card>
                    <CardHeader>使用方法</CardHeader>
                    <CardBody>
                        <div>
                            <p className={'mb-4'}>
                                查询所有粉丝量低于1000的主播
                                <Code className='ml-2'>`{`select * from ? where Fans < 1000`}`</Code>
                            </p>
                            <p className={'mb-4'}>
                                按粉丝/总督比 排序
                                <Code className='ml-2'>`{`select * from ? order by SUBSTRING(Guard,1,1)/Fans desc`}`</Code>
                            </p>
                            <p className={'mb-4'}>
                                查找签名包含妖精管理局的主播
                                <Code className='ml-2'>`{`select * from ? where Bio like '%妖精管理局%'`}`</Code>
                            </p>
                            <p className={'mb-4'}>
                                查找认证信息包含 [高能主播] 的主播，并按粉丝量升序排序
                                <Code className='ml-2'>`{`select * from ? where Verify like '%高能主播%' order by Fans`}`</Code>
                            </p>
                            <p className={'mb-4'}>
                                查找UID为504140200的主播
                                <Code className='ml-2'>`{`select * from ? where UID = 504140200`}`</Code>
                            </p>
                            <p className={'mb-4'}>
                                查找名字包含[兔]的主播
                                <Code className='ml-2'>`{`select * from ? where UName like '%兔%'`}`</Code>
                            </p>
                        </div>
                    </CardBody>
                </Card>}>
                    <Input className='max-w-xs mb-4 '
                           label={'高级筛选'} ref={rawSQLRef} isClearable onKeyDown={event => {
                        if (event.key === "Enter") {


                            try {
                                var start = new Date().getTime();
                                var query = alasql(event.target.value,[list])
                                addToast({
                                    title: "查询成功",
                                    description: `共找到${query.length}条记录，耗时${new Date().getTime()-start}ms`,
                                    color: 'success'
                                })
                                setFiltered(query);
                            } catch (e) {
                                addToast({
                                    title: "查询失败",
                                    description: `没有符合条件或语法错误`,
                                    color: 'danger'
                                })
                            }



                        }
                    }} onClear={() => {
                        setFiltered(list)
                    }}
                    ></Input>
                </Tooltip>
            </div>
            <Listbox
                virtualization={{
                    maxListboxHeight: calcHeight()-120,
                    itemHeight: 300,
                }}
                hideSelectedIcon
                variant={'light'}
                isVirtualized>
                {filted.slice(0,2000).map((item, index) => (
                    <ListboxItem key={index} value={item.value} css={{width: '100%'}} aria-label={item.label}
                                 textValue={''}
                                 onClick={() => {
                                     window.open(location.origin + "/liver/" + item.UID)
                                 }}
                    >
                        <LiverCard
                            Rank={index}
                            Avatar={`${protocol}://${host}:${port}${import.meta.env.PROD ? '' : '/api'}/face?mid=${item.UID}`}
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
                right: '20px',
                bottom: '20px',
                width: '180px',
                height: '60px',
            }}>
                <Input label="Search" onValueChange={(e) => {
                    setNameFilter(e)

                }}/>
            </div>

        </div>
    );

}

const LiverCard = memo(function LiverCard(props) {
    const up = props.DailyDiff >= 0
    return (
        <Card isHoverable style={{width: "100%", marginTop: "16px"}}>
            <CardBody style={{
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                width: "100%",
                padding: "16px"
            }}>


                <div style={{display: "flex", alignItems: "center"}}>
                    <Avatar
                        src={props.Avatar}
                        alt={props.UName}
                        width={64}
                        height={64}
                        radius="full"
                    />
                    <div style={{marginLeft: "12px"}}>
                        <p style={{fontSize: "16px", fontWeight: "500", margin: 0}}>{props.UName}</p>
                    </div>
                </div>
                <Chip style={{margin: '4px'}} onClick={() => {
                    toSpace(props.UID)
                }}>{props.UID}</Chip>


                <div style={{display: "flex", flexDirection: "column", alignItems: "center", lineHeight: 1.6}}>
                    <p style={{margin: 0}}>关注：{formatNumber(props.Fans)}</p>
                    <p style={{margin: 0}}>日增：{<span
                        style={{color: up ? '#00cc00' : '#ff0000'}}>{up ? '▲' : '▼'}</span>}{props.DailyDiff}</p>
                    <p style={{margin: 0}}>大航海：{props.Guard}</p>
                    <p style={{margin: 0, color: "#888"}}>上次直播：{formatTime(props.LastActive)}</p>
                    <p>{props.Bio}</p>
                    {props.Verify === '' ? <></> : <p style={{color: 'rgba(190,151,48,1)'}}>{props.Verify}</p>}
                    <p>Rank:{props.Rank + 1}</p>
                </div>

            </CardBody>
        </Card>
    )
})
export default ListPage;