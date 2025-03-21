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



export const CheckIcon = ({size, height, width, ...props}) => {
    return (
        <svg
            fill="none"
            height={size || height || 24}
            viewBox="0 0 24 24"
            width={size || width || 24}
            xmlns="http://www.w3.org/2000/svg"
            {...props}
        >
            <path
                d="M12 2C6.49 2 2 6.49 2 12C2 17.51 6.49 22 12 22C17.51 22 22 17.51 22 12C22 6.49 17.51 2 12 2ZM16.78 9.7L11.11 15.37C10.97 15.51 10.78 15.59 10.58 15.59C10.38 15.59 10.19 15.51 10.05 15.37L7.22 12.54C6.93 12.25 6.93 11.77 7.22 11.48C7.51 11.19 7.99 11.19 8.28 11.48L10.58 13.78L15.72 8.64C16.01 8.35 16.49 8.35 16.78 8.64C17.07 8.93 17.07 9.4 16.78 9.7Z"
                fill="currentColor"
            />
        </svg>
    );
};

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
                setCurrentStream(sort[0].Stream)
                setIsFirst(false);
            }
            setMonitor(sort)


        })
    }
    const refresh = () => {
        axios.get(`${protocol}://${host}:${port}/history?room=${room}&last=${lastRef.current}`).then(res => {
            if (res.data.data.length != 0) {
                setMessage(res.data.data)
                setPartMessages(res.data.data)
                res.data.data.map((item) => {
                    if (item.ActionName !== 'enter') {
                        message.push(item)
                    }
                })

                setMessage(message)
                setLast(res.data.data[res.data.data.length - 1].UUID)
            }

        })
    }

    useEffect(() => {
        initRoomList();
    }, [])

    const lastRef = React.useRef(null);

    useEffect(() => {
        lastRef.current = last;
    }, [last]);

    useEffect(() => {
        setLast("")
        setMessage([])
        initRoomList()
        refresh();
        console.log("useEffect")

        const interval = setInterval(() => {
            initRoomList()
            refresh();

        }, 1000);

        return () => clearInterval(interval);
    }, [room]);


    const chatRef = React.useRef(null);
    const host = location.hostname;

    const port = debug ? 8080 : location.port;

    const protocol = location.protocol.replace(":", "")

    function GiftPart(props) {
        return (
            <div className={classes.giftArea}>
                <p>{props.name}</p>
                <img src={`${protocol}://${host}:${port}/proxy?url=${props.img}`} alt="" />
            </div>
        );
    }
    const rowVirtualizer = useVirtualizer({
        count: message.length,
        getScrollElement: () => chatRef.current, // 绑定滚动容器
        estimateSize: () => 160, // 预估行高
        overscan: 30
    });
    useEffect(() => {
        rowVirtualizer.scrollToIndex(message.length - 1, { align: "end" });
    }, [message.length]);
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
                                                <Avatar src={`${protocol}://${host}:${port}/proxy?url=${item.Face}`}
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
                <div className={classes.chatColumn}>
                    <div ref={chatRef} style={{ height: "100%", width: "100%", overflow: "auto" }}>
                        <div style={{ height: `${rowVirtualizer.getTotalSize()}px`, position: "relative" }}>
                            {rowVirtualizer.getVirtualItems().map(virtualRow => {
                                const item = message[virtualRow.index];
                                return (
                                    <div
                                        key={item.UUID}
                                        ref={virtualRow.measureElement}
                                        style={{
                                            position: "absolute",
                                            top: 0,
                                            left: 0,
                                            width: "100%",
                                            transform: `translateY(${virtualRow.start}px)`,
                                        }}
                                    >
                                        <Card style={{ margin: "15px" }} isHoverable>
                                            <CardBody>
                                                <div style={{ display: "flex" }}>
                                                    <Avatar
                                                        src={`${protocol}://${host}:${port}/proxy?url=${item.Face}`}
                                                        onClick={() => toSpace(UID)}
                                                    />
                                                    <div>
                                                        <div className={classes.nameRow}>
                                                            <p>{item.FromName}</p>
                                                            {item.MedalName && (
                                                                <Chip
                                                                    startContent={<CheckIcon size={18} />}
                                                                    variant="faded"
                                                                    style={{ marginLeft: "8px", background: item.MedalColor,color:'white'}}
                                                                >
                                                                    {item.MedalName}
                                                                    <span className="ml-2 text-xs font-bold px-2 py-0.5 rounded-full">
                                                            {item.MedalLevel}
                                                        </span>
                                                                </Chip>
                                                            )}
                                                        </div>
                                                        <div style={{ marginLeft: "8px" }}>
                                                            {item.ActionName === "msg" ? (
                                                                <span className="messageText" >{item.Extra}</span>
                                                            ) : (
                                                                <GiftPart name={item.GiftName} img={item.GiftPicture} />
                                                            )}
                                                        </div>
                                                    </div>
                                                </div>
                                            </CardBody>
                                            <CardFooter className={classes.msgTime}>
                                                <p>
                                                    {String(new Date(item.CreatedAt).getHours()).padStart(2, '0')}:
                                                    {String(new Date(item.CreatedAt).getMinutes()).padStart(2, '0')}
                                                </p>
                                            </CardFooter>
                                        </Card>
                                    </div>
                                );
                            })}
                        </div>
                    </div>

                </div>

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

                        </Tab>
                        <Tab title="大航海" >

                        </Tab>
                    </Tabs>
            </div>

        </div>
    );
}

export default ChatPage;