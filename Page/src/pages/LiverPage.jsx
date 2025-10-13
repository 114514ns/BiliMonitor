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
    ModalHeader, Select, SelectItem, Switch,
    Tooltip
} from "@heroui/react";
import HoverMedals from "../components/HoverMedals";
import {FansList} from "../components/RankDialog";
import LiveStatisticCard from "../components/LiveStatisticCard";
import {AnimatePresence,motion} from "framer-motion";
import {FansChart, GuardChart} from "../components/LineChart";



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
    }, [noDM,month])



    const [diffMode,setDiffMode] = React.useState(false);
    return (
        <div>
            <Modal isOpen={open} onOpenChange={() => {
                setOpen(!open)
            }}    className={`grid transition-[grid-template-rows] duration-300 ease-out ${
                diffMode ? 'grid-rows-[1fr]' : 'grid-rows-[0fr] h-[70vh]'
            }`}>
                <ModalContent>
                    <ModalHeader className="flex flex-col gap-1">
                        {guardTime}
                    </ModalHeader>
                    <ModalBody>
                        <div>
                            <Switch isSelected={diffMode}  onValueChange={(e => {
                                setDiffMode(e)
                                if (!e) {
                                    setGuard(orig)
                                }
                            })}>Diff</Switch>
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
                        <FansList fans={guard} height={800} onItemClick={(e) => {
                            console.log(e)
                            redirect('/user/' + e.UID)
                        }}/>
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
                    className="rounded-xl bg-green-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">粉丝<br/>
                    <span className="font-semibold">{parseInt(space.Fans).toLocaleString()}</span>
                </div>
                <div
                    className="rounded-xl bg-yellow-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">大航海<br/>
                    <span
                        className="font-semibold">{space.Guard}</span>
                </div>
                <Tooltip content={<HoverMedals ruid={id}/>} delay={1500}>
                    <div
                        onContextMenu={e => {

                        }}
                        className="rounded-xl bg-pink-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">粉丝牌<br/>
                        <span
                            className="font-semibold">{space.Medal}</span>
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
                            setGuard(response.data.data);
                            setOrig(response.data.data)
                            setOpen(true)
                        })
                    }}/>
                </div>
            </div>
            <Switch onValueChange={(value) => {
                setNoDM(value)
            }} defaultSelected={false}>
                显示所有场次
            </Switch>
            <div className={'grid grid-cols-1 md:grid-cols-4 2xl:grid-cols-5'}>
                {lives.map((live, index) => (
                    <LiveStatisticCard item={live} showUser={false} key={live.ID}/>
                ))}
            </div>
        </div>
    );
}



export default LiverPage;