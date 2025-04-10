import React, {useEffect, useMemo, useState} from 'react';
import classes from "./ChatPage.module.css";
import axios from "axios";
import {
    Autocomplete, AutocompleteItem,
    Avatar,
    Badge,
    Button,
    Card,
    CardBody,
    CardFooter,
    Chip,
    Input,
    Tab,
    Tabs,
    Tooltip
} from "@heroui/react";
import ReactPlayer from 'react-player/streamable'
import WatcherList from "../components/WatcherList";
import ChatArea from "../components/ChatArea";



export const CheckIcon = React.memo(({ size = 24, color = "currentColor", ...props }) => {
    return (
        <svg width={size} height={size} {...props}>
            <use href="#icon-check" fill={color} />
        </svg>
    );
});

function ChatPage(props) {




    const [room, setRoom] = useState("");



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
            const n = []
            sort.forEach(live => {
                live.OnlineWatcher = []
                live.GuardList = []
                n.push(live)
            })
            if (JSON.stringify(n) !== JSON.stringify(monitor)) {
                setMonitor(sort)
            }




        })
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

    useEffect(() => {
        console.log("stream changed");
    },[currentStream])


    const chatRef = React.useRef(null);
    const host = location.hostname;

    const port = location.port;

    const protocol = location.protocol.replace(":", "")

    const users = [
        {
            id: 1,
            name: "Tony Reichert",
            role: "CEO",
            team: "Management",
            status: "active",
            age: "29",
            avatar: "https://d2u8k2ocievbld.cloudfront.net/memojis/male/1.png",
            email: "tony.reichert@example.com",
        },
        ]


    return (
        <div className={classes.root}>

            <div className={classes.roomColumn}>
                <Autocomplete
                    className="max-w-xs"
                    defaultItems={users}
                    label="Assigned to"
                    labelPlacement="inside"
                    placeholder="Select a user"
                    variant="bordered"
                >
                    {(user) => (
                        <AutocompleteItem key={user.id} textValue={user.name}>
                            <div className="flex gap-2 items-center">
                                <Avatar alt={user.name} className="flex-shrink-0" size="sm" src={user.avatar} />
                                <div className="flex flex-col">
                                    <span className="text-small">{user.name}</span>
                                    <span className="text-tiny text-default-400">{user.email}</span>
                                </div>
                            </div>
                        </AutocompleteItem>
                    )}
                </Autocomplete>
                {monitor.map(item => {
                    return (
                        <div onClick={() => {
                            console.log(item.LiveRoom)
                            setRoom(item.LiveRoom)
                            setCurrentStream(item.Stream)
                        }} key={item.UID}>
                            <Card isHoverable={true} style={{ margin: "10px", width: "95%" }} radius={'none'}>
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
                        <Tab title={`在线：${monitor.filter((e) =>e.LiveRoom == room)[0]?.OnlineCount}`}>

                            <WatcherList room={room} type={'online'}/>
                        </Tab>

                        <Tab title={`大航海：${monitor.filter((e) =>e.LiveRoom == room)[0]?.GuardCount}`} >
                            <WatcherList room={room} type={'guard'}/>
                        </Tab>
                    </Tabs>
            </div>

        </div>
    );
}

export default ChatPage;