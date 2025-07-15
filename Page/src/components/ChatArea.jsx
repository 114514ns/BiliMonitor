import React, {useEffect, useRef, useState} from 'react';
import classes from "../pages/ChatPage.module.css";
import {Avatar, Card, CardBody, CardFooter, Chip, Tooltip} from "@heroui/react";
import {CheckIcon} from "../pages/ChatPage";
import axios from "axios";
import {Listbox, ListboxItem} from "@heroui/react";
import { motion } from 'framer-motion';


function GiftPart(props) {
    return (
        <div className={classes.giftArea}>
            <Tooltip content={props.name}>
                <img src={props.img} alt="" style={{objectFit: "cover",width:"60px"}} />
            </Tooltip> * {props.amount}  CNY {props.amount*props.price}
        </div>
    );
}
function calcHeight() {
    const vh = window.innerHeight;
    const rem = parseFloat(getComputedStyle(document.documentElement).fontSize);
    const result = vh - 4 * rem;
    return result*0.9; // 返回的是 px 数值
}
function parseMessage(message, emojiMap) {
    const regex = /\[([^\[\]]+)\]/g;
    const parts = [];
    let lastIndex = 0;
    let match;

    while ((match = regex.exec(message)) !== null) {
        const emojiKey = match[0]; // 包括中括号的键
        const matchIndex = match.index;

        if (matchIndex > lastIndex) {
            parts.push(message.slice(lastIndex, matchIndex));
        }

        if (emojiMap[emojiKey]) {
            parts.push(
                `<img src="${emojiMap[emojiKey]}" alt="${emojiKey}" style="width: 20px; height: 20px; vertical-align: text-bottom; display: inline-block;">`
            );
        } else {
            parts.push(emojiKey);
        }

        lastIndex = regex.lastIndex;
    }

    if (lastIndex < message.length) {
        parts.push(message.slice(lastIndex));
    }

    return parts.join('');
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
    const isAtBottom = (element) => {
        const threshold = 50;
        return (
            element.scrollTop + element.clientHeight >= element.scrollHeight - threshold
        );
    };
    useEffect(() => {
        if (chatRef.current) {
            const element = chatRef.current.children[0].children[0].children[0]
            element.scrollTop=99999999;
            console.log("top ",element.scrollTop,"height ",element.height);


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
                                    <motion.div style={{width:'100%',display:'flex',justifyContent:'center'}}
                                                initial={{ opacity: 0, y: 30 }}
                                                animate={{ opacity: 1, y: 0 }}
                                                exit={{ opacity: 0, y: 30 }}
                                                transition={{ duration: 0.4 }}
                                    >
                                        <ChatItem item={item} />
                                    </motion.div>
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
                                <span
                                    className="messageText"
                                    dangerouslySetInnerHTML={{ __html: parseMessage(item.Extra, item.Emoji) }}
                                ></span>
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