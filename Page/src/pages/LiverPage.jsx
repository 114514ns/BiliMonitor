import React, {useState} from 'react';
import {useNavigate, useParams} from "react-router-dom";
import axios from "axios";
import {
    addToast,
    Avatar, Button,
    Card,
    CardBody,
    CardHeader,
    Image,
    Modal,
    ModalBody,
    ModalContent,
    ModalHeader, Pagination, Select, SelectItem, Switch,
    Tooltip
} from "@heroui/react";
import HoverMedals from "../components/HoverMedals";
import {FansList} from "../components/RankDialog";
import LiveStatisticCard from "../components/LiveStatisticCard";
import {AnimatePresence,motion} from "framer-motion";
import {FansChart, GuardChart} from "../components/LineChart";
import DynamicCard from "../components/DynamicCard";

function calcValid(array) {
    var count = 0
    array.forEach(element => {
        if (!inspectGuard(element)) {
            count++
        }
    })

    return count
}

function parseCharge(str) {
    var sum = 0
    JSON.parse(str).forEach((item) => {
        sum = sum + item.Count*item.Price
    })
    return sum
}

function LiverPage(props) {
    const [fansChart, setFansChart] = React.useState([]);

    const [guardChart, setGuardChart] = React.useState([]);

    const [space, setSpace] = React.useState({});

    const [lives, setLives] = React.useState([]);


    const [open, setOpen] = React.useState(false);

    const [guard, setGuard] = React.useState([]);

    const [guardTime,setGuardTime] = React.useState("");

    const redirect = useNavigate();

    const [orig,setOrig] = React.useState([]);

    const [noDM, setNoDM] = React.useState(false);

    const [month,setMonth] = useState(3)

    const [stream,setStream] = React.useState({})

    const [guild,setGuild] = React.useState("");

    const [guardId,setGuardId]  = React.useState('')

    const [load,setLoad] = React.useState(false)

    const [dynCount,setDynCount] = React.useState(0)

    const listRef = React.useRef()

    const [showAmount,setShowAmount] = React.useState(false);


    let {id} = useParams();

    React.useEffect(() => {
        if (localStorage.getItem("guild")) {
            JSON.parse(localStorage.getItem("guild")).forEach((item) => {
                if (item.uid === parseInt(id)) {
                    setGuild(item.guild_name)
                }
            })
        }

    },[])

    React.useEffect(() => {

        axios.get(`${protocol}://${host}:${port}/api/chart/fans?uid=${id}&month=${month}`).then((response) => {
            setFansChart(response.data.data??[]);
        })
        axios.get(`${protocol}://${host}:${port}/api/liver/space?uid=${id}`).then((response) => {
            setSpace(response.data);
            document.title = response.data.UName + '的直播记录'
        })
        axios.get(`${protocol}://${host}:${port}/api/chart/guard?uid=${id}&month=${month}`).then((response) => {
            var dst = []
            var map = [19998, 1998, 138]
            response.data.data?.forEach(element => {
                dst.push({
                    UpdatedAt: element.UpdatedAt,
                    Guard: element.Guard.split(",").reduce((a, b) => parseInt(a) + parseInt(b)),
                    ID: element.ID,
                })

            })
            dst.sort((a,b) => a.Level > b.Level)
            setGuardChart(dst??[])
        })
        axios.get(`${protocol}://${host}:${port}/api/live?uid=${id}&limit=1000&no_dm=${noDM}`).then((response) => {
            setLives(response.data.lives);
        })

        axios.get("/api/dynamics/count?mid=" + id).then((response) => {
            setDynCount(response.data.count)
        })
    }, [noDM,month])

    const [diffMode,setDiffMode] = React.useState(false);

    const [inspectMode,setInspectMode] = React.useState(false);

    const [showDyn,setShowDyn] = React.useState(false);

    const [dynList,setDynList] = React.useState([])

    const [dynPage,setDynPage] = React.useState(1)

    React.useEffect(() => {
        axios.get(`/api/guard?id=${guardId}&inspect=true`).then((response) => {
            var t = response.data.data
            t.sort((a,b) => b.Level-a.Level)
            setGuard(t)
            setOrig(t)
            if (inspectMode) {
                setLoad(true)
            }
        })
    },[inspectMode])

    const DYN_SIZE = 50



    return (
        <div>
            <Modal isOpen={open} onOpenChange={() => {
                if (open === true) {
                    setInspectMode(false);
                    setLoad(false)
                }
                setOpen(!open)
            }}    className={` grid transition-[grid-template-rows] duration-300 ease-out ${
                diffMode ? 'grid-rows-[1fr]' : 'grid-rows-[0fr] h-[70vh]'
            }`}>
                <ModalContent>
                    <ModalHeader className="flex flex-col gap-1">
                        <span>{guardTime}</span>
                        {inspectMode &&  load && <span>{inspectMode && `${calcValid(guard)} / ${guard.length}`}</span>}
                    </ModalHeader>
                    <ModalBody>
                        <div>
                            <div className={'flex flex-col'}>
                                <Switch isSelected={diffMode}  onValueChange={(e => {
                                    setDiffMode(e)
                                    if (!e) {
                                        setGuard(orig)
                                    }
                                })} isDisabled={inspectMode}>Diff</Switch>
                                <Switch isSelected={inspectMode}  onValueChange={(e => {
                                    setInspectMode(e)
                                })} className={'mt-2'} isDisabled={diffMode}>Inspect</Switch>
                            </div>
                            <AnimatePresence>
                                {diffMode && (
                                    <motion.div
                                        key="select"
                                        initial={{ opacity: 0, y: -5 }}
                                        animate={{ opacity: 1, y: 0 }}
                                        exit={{ opacity: 0, y: -5 }}
                                        transition={{ duration: 0.2 }}
                                    >
                                        <Select className="mt-2">
                                            {guardChart.filter(e => new Date(e.UpdatedAt).getTime() > new Date(guardTime).getTime()).map(e => {
                                                var str = new Date(e.UpdatedAt).toLocaleString()
                                                return (
                                                    <SelectItem key={str} value={str} onPress={() => {
                                                        axios.get(`${protocol}://${host}:${port}/api/guard?id=${e.ID}`).then((response) => {
                                                            var dst = response.data.data
                                                            const oldIds = new Set(orig.map(item => item.UID));
                                                            const newIds = new Set(dst.map(item => item.UID));
                                                            const added = dst
                                                                .filter(item => !oldIds.has(item.UID))
                                                                .map(item => ({ ...item, Label: 'add' }));

                                                            const removed = orig
                                                                .filter(item => !newIds.has(item.UID))
                                                                .map(item => ({ ...item, Label: 'remove' }));
                                                            setGuard([...added, ...removed]);

                                                        })
                                                    }} aria-label={''}>{str}
                                                    </SelectItem>
                                                )
                                            })}
                                        </Select>
                                    </motion.div>
                                )}
                            </AnimatePresence>
                        </div>
                        <FansList fans={guard} height={600} onItemClick={(e) => {
                            console.log(e)
                            redirect('/user/' + e.UID)
                        }} inspect={inspectMode}/>
                    </ModalBody>
                </ModalContent>
            </Modal>
            <Modal isOpen={showDyn} onOpenChange={() => {
                setShowDyn(!showDyn)
            }}    className={'overflow-x-hidden'} scrollBehavior={'inside'}>
                <ModalContent>
                    <ModalHeader className="flex flex-col gap-1">
                        <span>{space.UName}的历史动态</span>
                    </ModalHeader>
                    <ModalBody className={'overflow-x-hidden'}>
                        <div ref={listRef}>
                            {dynList.slice((dynPage-1)*DYN_SIZE,dynPage*DYN_SIZE).map((item,i)=>{
                                return (
                                    <DynamicCard item={item} key={i} onClick={() => {
                                        window.open("https://t.bilibili.com/" + item.IDStr)
                                    }}/>
                                )
                            })}
                            <Pagination initialPage={1} total={Math.ceil(dynList.length/DYN_SIZE)} onChange={(e) => {
                                setDynPage(e)
                                listRef.current.parentElement.scrollTo({
                                    top: 0,
                                    behavior: "smooth"
                                });
                            }}/>;
                        </div>
                    </ModalBody>
                </ModalContent>
            </Modal>
            <div className={'flex flex-col mb-6 sm:justify-center items-center '}>
                <Avatar src={`${AVATAR_API}${id}`}
                        size={'lg'} onClick={() => {
                    toSpace(id)
                }}/>
                <div className={'ml-4 flex flex-col items-center'}>
                    <span className={'font-bold'}>{space.UName}</span>
                    {space.Verify &&
                        <div className={'flex items-center'}>
                            <svg style={{color: '#ffcc00'}} width="16" height="16" viewBox="0 0 16 16" fill="none"
                                 xmlns="http://www.w3.org/2000/svg">
                                <path
                                    d="M16 8C16 12.4183 12.4183 16 8 16C3.58172 16 0 12.4183 0 8C0 3.58172 3.58172 0 8 0C12.4183 0 16 3.58172 16 8Z"
                                    fill="currentColor"></path>
                                <path fill-rule="evenodd" clip-rule="evenodd"
                                      d="M7.28832 12.7244C7.20127 12.767 7.1148 12.7988 7.02538 12.8C6.80863 12.8042 6.62919 12.6296 6.62564 12.4101C6.62742 12.3717 6.63512 12.3351 6.64814 12.2997L7.40676 8.78586L4.26392 8.79186C4.03651 8.79545 3.85825 8.6209 3.85352 8.40196C3.85588 8.27299 3.9228 8.15362 4.03118 8.08524L8.72206 3.2901C8.80852 3.23732 8.90149 3.20133 8.99743 3.20013C9.21477 3.19653 9.39303 3.37108 9.39776 3.59063C9.39599 3.65541 9.37822 3.71839 9.34506 3.77358L8.59118 7.23047H11.7388C11.9614 7.22687 12.1403 7.40142 12.1444 7.62096C12.1426 7.75113 12.0757 7.8705 11.9668 7.93888L7.28832 12.7244Z"
                                      fill="white"></path>
                            </svg>
                            <span className={'font-bold ml-2'}>{space.Verify}</span>

                        </div>
                    }
                    {guild && <span className={'font-thin text-sm'}>公会：{guild}</span>}
                    <span className={'font-thin text-sm'}>{space.Bio}</span>
                </div>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-2 text-sm">
                <div
                    className="rounded-xl bg-green-50 dark:bg-gray-700 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">粉丝<br/>
                    <span className="font-semibold">{parseInt(space.Fans).toLocaleString()}</span>
                </div>
                <div
                    className="rounded-xl bg-yellow-50 dark:bg-gray-600 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">大航海<br/>
                    <span
                        className="font-semibold">{space.Guard}</span>
                </div>
                <Tooltip content={<HoverMedals ruid={id}/>} delay={1500}>
                    <div
                        onContextMenu={e => {

                        }}
                        className="rounded-xl bg-pink-50 dark:bg-gray-400 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">粉丝牌<br/>
                        <span
                            className="font-semibold">{space.Medal}</span>
                    </div>
                </Tooltip>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-sm mt-4">
                <div
                    onClick={() => {
                        setShowDyn(true)
                        const start = new Date().getTime()
                        axios.get("/api/dynamics?mid=" + id).then(res=>{
                            console.log(new Date().getTime()-start) ;
                            var array = res.data.data
                            var m = new Map()
                            array.forEach(element => {
                                m.set(element.ID, element)
                            })
                            console.log(new Date().getTime()-start) ;
                            array.forEach(element => {
                                if (element.ForwardFrom !== 0) {
                                    element.ForwardDynamic = m.get(element.ForwardFrom)
                                }
                            })
                            console.log(new Date().getTime()-start) ;

                            var newArray = []
                            var v0 = parseInt(id)
                            array.forEach(element => {
                                if (element.UID === v0) {
                                    newArray.push(element)
                                }
                            })
                            console.log(new Date().getTime()-start) ;
                            setDynList(newArray)
                            console.log(new Date().getTime()-start) ;
                        })
                    }}
                    className="rounded-xl bg-blue-50 dark:bg-gray-500 p-2 transition-transform duration-200 hover:scale-102 hover:shadow-lg ">动态<br/>
                    <span className="font-semibold">{parseInt(dynCount).toLocaleString()}</span>
                </div>
                <Tooltip content={<div>
                    <p>直播收入 {parseInt(space.Amount).toLocaleString()}</p>
                    {space && space.Charge && space.Charge!=='null' && <p>充电收入：{parseCharge(space.Charge).toLocaleString()}</p>}
                </div>} isOpen={showAmount} onOpenChange={(o) => {
                    setShowAmount(o)
                }}>
                    <div
                        onContextMenu={() => {
                            setShowAmount(!showAmount)
                        }}
                        className="rounded-xl bg-yellow-50 dark:bg-gray-500 p-2 transition-transform duration-200 ">30日内流水<br/>
                        <span
                            className="font-semibold">{(((space && space.Charge && space.Charge!== 'null')?parseCharge(space.Charge):0 )+ parseInt(space.Amount)).toLocaleString()}</span>
                    </div>
                </Tooltip>
            </div>
            <div>
                <Button onClick={e => {
                    setMonth(month+3)
                }} className={'mt-2'}>更多</Button>
            </div>
            <div className={'grid grid-cols-1 sm:grid-cols-2 w-full'}>
                <div className="h-[300px]">
                    <FansChart data={fansChart}/>
                </div>
                <div className="h-[300px]">
                    <GuardChart data={guardChart} onClick={(index) => {
                        setGuardTime(new Date(guardChart[index].UpdatedAt).toLocaleString())
                        axios.get(`/api/guard?id=${guardChart[index].ID}`).then((response) => {
                            var t = response.data.data
                            t.sort((a,b) => b.Level-a.Level)
                            setGuard(t)
                            setOrig(t)
                            setOpen(true)
                            setGuardId(guardChart[index].ID)
                        })
                    }}/>
                </div>
            </div>
            <Switch onValueChange={(value) => {
                setNoDM(value)
            }} defaultSelected={false}>
                显示所有场次
            </Switch>
            <div className="grid 3xl:grid-cols-5 2xl:grid-cols-4">

                {lives.map((live, index) => (
                    <LiveStatisticCard item={live} showUser={false} key={live.ID}/>
                ))}
            </div>
        </div>
    );
}



export default LiverPage;