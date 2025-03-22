import React, {useEffect, useMemo, useState} from 'react';
import {useParams} from "react-router-dom";
import classes from "./ChatPage.module.css";
import axios from "axios";
import {Avatar, Badge, Button, Card, CardBody, CardFooter, Chip, Input, Tab, Tabs} from "@heroui/react";
import { motion, AnimatePresence } from "framer-motion";
import {AutoSizer, List} from 'react-virtualized';
import {useVirtualizer} from "@tanstack/react-virtual";
import {Tooltip} from "recharts";
import {Slider} from "@heroui/slider";
import { Volume2, VolumeX, Volume1 } from "lucide-react";
import ReactPlayer from "react-player";
import WatcherList from "../components/WatcherList";
import ChatArea from "../components/ChatArea";



export const CheckIcon = React.memo(({ size = 24, color = "currentColor", ...props }) => {
    return (
        <svg width={size} height={size} {...props}>
            <use href="#icon-check" fill={color} />
        </svg>
    );
});

function MuteButton() {
    const [volume, setVolume] = useState(50);
    const [isMuted, setIsMuted] = useState(false);
    const [showSlider, setShowSlider] = useState(false);

    const toggleMute = () => {
        setIsMuted(!isMuted);
    };

    const handleVolumeChange = (val) => {
        setVolume(val);
        setIsMuted(val === 0);
    };

    return (
        <div
            className=""
            onMouseEnter={() => setShowSlider(true)}
            onMouseLeave={() => setShowSlider(false)}
        >
            {showSlider && (
                <Slider
                    value={isMuted ? 0 : volume}
                    onChange={handleVolumeChange}
                    min={0}
                    max={100}
                    step={1}
                    style={{position: 'absolute',width:'20%',marginBottom:'20px',marginLeft:'20px'}}
                />
            )}
            {isMuted ? <VolumeX /> : volume > 50 ? <Volume2 /> : <Volume1 />}
        </div>
    );
}


function ChatPage(props) {


    const [message, setMessage] = useState([]);

    const [partMessages, setPartMessages] = useState([]);

    const [room, setRoom] = useState("");

    const [last, setLast] = useState("");

    const [currentStream, setCurrentStream] = useState("");



    const [monitor, setMonitor] = useState([])

    var [isFirst, setIsFirst] = useState(true);

    const initRoomList = () => {
        axios.get(`${protocol}://${host}:${port}/monitor`).then(res => {
            const sort = res.data.lives.sort((a, b) => {
                if (a.Live === b.Live) {
                    return a.UID > b.UID ? 1 : -1;
                }
                return a.Live ? -1 : 1;
            });
            if (isFirst) {
                setRoom(sort[0].LiveRoom)
                console.log(sort[0].LiveRoom)
                setCurrentStream(sort[0].Stream)
                setIsFirst(false);
            }
            if (JSON.stringify(sort) !== JSON.stringify(monitor)) {
                setMonitor(sort)
            }




        })
    }

    const getUser = ()=> {
        if (monitor.filter((e) => e.LiveRoom === room).length === 0) return []
        return monitor.filter((e) => e.LiveRoom === room)[0].OnlineWatcher
    }
    const getGuard = ()=> {
        if (monitor.filter((e) => e.LiveRoom === room).length === 0) return []
        return monitor.filter((e) => e.LiveRoom === room)[0].GuardList
    }

    useEffect(() => {
        initRoomList();
    }, [])



    useEffect(() => {
        initRoomList()
        console.log("room changed")

        const interval = setInterval(() => {
            initRoomList()

        }, 1000);

        return () => clearInterval(interval);
    }, [room]);


    const chatRef = React.useRef(null);
    const host = location.hostname;

    const port = debug ? 8080 : location.port;

    const protocol = location.protocol.replace(":", "")




    return (
        <div className={classes.root}>

            <div className={classes.roomColumn}>
                {monitor.map(item => {
                    return (
                        <div onClick={() => {
                            console.log(item.LiveRoom)
                            setRoom(item.LiveRoom)
                            setMessage([])
                            setCurrentStream(item.Stream)
                        }} key={item.UID}>
                            <Card isHoverable={true} style={{ margin: "10px", width: "100%" }}>
                                <CardBody>
                                    <div style={{
                                        display: "flex",
                                        flexDirection: "row",
                                        alignItems: "center",
                                        justifyContent: "space-between",
                                        flexWrap: "wrap"
                                    }}>
                                        <div style={{ display: "flex", alignItems: "center", gap: "8px", minWidth: 0 }}>
                                            <Badge color={item.Live ? "success" : "default"} content="">
                                                <Avatar src={item.Face}
                                                        onClick={() => toSpace(item.UID)} />
                                            </Badge>
                                            <div style={{ minWidth: 0 }}>
                                                <p style={{
                                                    margin: 0,
                                                    fontSize: "16px",
                                                    fontWeight: "bold",
                                                    whiteSpace: "nowrap",
                                                    overflow: "hidden",
                                                    textOverflow: "ellipsis",
                                                    maxWidth: "120px"
                                                }}>
                                                    {item.UName}
                                                </p>
                                                <p style={{
                                                    margin: 0,
                                                    fontSize: "14px",
                                                    color: "#888",
                                                    whiteSpace: "nowrap",
                                                    overflow: "hidden",
                                                    textOverflow: "ellipsis",
                                                    maxWidth: "180px"
                                                }}>
                                                    {item.Title}
                                                </p>
                                            </div>
                                        </div>
                                    </div>
                                </CardBody>
                            </Card>


                        </div>
                    )
                })}
            </div>
            <div className={classes.chat}>


                {room && <ChatArea room={room} />}
                <Input label="Email" type="email"  endContent={
                    <Button
                        size="sm"
                        color="primary"
                        onPress={() => console.log("发送消息")}
                    >
                        发送
                    </Button>
                }/>
            </div>
            <div className={classes.right}>
                <ReactPlayer url={currentStream} controls={true} playing={true}/>

                    <Tabs aria-label="Options" style={{ marginTop: "10px" ,width: "100%",display:'flex',justifyContent:'space-between' }} fullWidth={true}>
                        <Tab title="在线">

                            <WatcherList list={getUser()}/>
                        </Tab>
                        <Tab title="大航海" >
                            <WatcherList list={getGuard()}/>
                        </Tab>
                    </Tabs>
            </div>

        </div>
    );
}

export default ChatPage;