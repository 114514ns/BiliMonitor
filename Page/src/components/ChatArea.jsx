import React, {useEffect, useRef, useState} from 'react';
import classes from "../pages/ChatPage.module.css";
import {Avatar, Card, CardBody, CardFooter, Chip} from "@heroui/react";
import {CheckIcon} from "../pages/ChatPage";
import axios from "axios";
import {Listbox, ListboxItem} from "@heroui/react";
function GiftPart(props) {
    return (
        <div className={classes.giftArea}>
            <p>{props.name}*{props.amount}：CNY {props.price}</p>
            <img src={props.img} alt="" style={{objectFit: "cover",width:"70px"}} />
        </div>
    );
}
function calcHeight() {
    const vh = window.innerHeight;
    const rem = parseFloat(getComputedStyle(document.documentElement).fontSize);
    const result = vh - 4 * rem;
    return result*0.9; // 返回的是 px 数值
}
function ChatArea(props) {
    const chatRef = useRef(null);
    const [last, setLast] = useState("");
    const [message, setMessage] = useState([]);
    const lastRef = useRef(null);
    const host = location.hostname;
    const port = location.port;
    const protocol = location.protocol.replace(":", "");

    useEffect(() => {
        lastRef.current = last;
    }, [last]);


    const refresh = () => {
        axios
            .get(`${protocol}://${host}:${port}/api/history?room=${props.room}&last=${lastRef.current}`)
            .then((res) => {
                if (res.data.data.length !== 0) {
                    const newData = res.data.data.filter((item) => item.ActionName !== "enter");

                    if (newData.length > 0) {
                        setMessage((prev) => {
                            const merged = [...prev, ...newData];
                            return JSON.stringify(merged) === JSON.stringify(prev) ? prev : merged;
                        });

                        setLast(res.data.data[res.data.data.length - 1].UUID);
                    }
                }
            });
    };

    useEffect(() => {
        if (chatRef.current) {
            chatRef.current.children[0].children[0].children[0].scrollTop=99999999;
            console.log("scroll");

        }
    }, [last]);

    useEffect(() => {
        setLast("");
        setMessage([]);
        refresh();

        const interval = setInterval(() => {
            refresh();
        }, 1000);

        return () => {
            clearInterval(interval);
            setMessage([]);
        };
    }, [props.room]);

    return (
        <div style={{height:'90%'}}>
            <div className={classes.chatColumn}>
                <div style={{ alignItems:'center'}} ref={chatRef}>
                        <Listbox
                            isVirtualized
                            autoFocus={'first'}
                            label={"Select from 1000 items"}
                            placeholder="Select..."
                            hideSelectedIcon
                            variant={'light'}
                            virtualization={{
                                maxListboxHeight: calcHeight(),
                                itemHeight: 160,
                            }}
                        >
                            {message.map((item, index) => (
                                <ListboxItem key={index} value={item.value} textValue={'1'} >
                                    <div style={{width:'100%'}}>
                                        <ChatItem item={item} />
                                    </div>
                                </ListboxItem>
                            ))}
                        </Listbox>
                </div>
            </div>
        </div>
    );
}
function ChatItem(props) {

    var item = props.item;
    var virtualRow = props.row
    return (
        <Card style={{width: "90%",marginTop:"10px"}}
              isHoverable
              key={`${item.UUID}`}

        >
            <CardBody>
                <div style={{ display: "flex" }}>
                    <img
                        src={item.Face}
                        style={{width:'40px',height:'40px',borderRadius:'50%'}}
                        onClick={() => toSpace(item.FromId)}
                    />
                    <div>
                        <div className={classes.nameRow}>
                            <p>{item.FromName}</p>
                            {item.MedalName && (
                                <Chip
                                    startContent={<CheckIcon size={18} />}
                                    variant="faded"
                                    style={{ background: item.MedalColor,color:'white',marginLeft:'8px'}}
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
                                <GiftPart name={item.GiftName} img={item.GiftPicture} price={item.GiftPrice.Float64} amount={item.GiftAmount.Int16}/>
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
    )
}

export default ChatArea;