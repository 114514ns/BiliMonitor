import React, {useEffect, useState} from 'react';
import {Button, Modal, ModalBody, ModalContent, ModalFooter, ModalHeader} from "@heroui/react";
import Markdown from "react-markdown";
import axios from "axios";
import remarkGfm from 'remark-gfm'
import "github-markdown-css/github-markdown-light.css";
import rehypeRaw from "rehype-raw";
import DynamicCard from "./DynamicCard";
import BVPlayer from "./BVPlayer";

function NoticeDialog(props) {


    const [seconds,setSeconds] = useState(0);
    /*

    useEffect(() => {
        var ref = setInterval(()=>{
            if (seconds <=0) {
                clearInterval(ref);
            }
            setSeconds((prev) => prev - 1);
        },1000)

        return () => {
            clearInterval(ref);
        }
    })

     */

    return (
        <div data-nosnippet>
            <Modal
                onClose={seconds <= 0 && props.onClose}
                isOpen={true}
                size={isMobile() ? '2xl' : '2xl'}
                // 2. 移除这里普通的 className
                // className={'max-h-2/3 sm:h-[75vh]'}

                // 3. 使用 classNames (插槽) 精确控制对话框本体 (base) 和 主体 (body)
                classNames={{
                    base: "sm:max-h-[75vh] sm:h-[75vh]", // 强制对话框高度为 75vh
                    body: "p-4" // 你可以根据需要调整内部 padding
                }}

                scrollBehavior={'inside'}
                backdrop={'blur'}
                isDismissable={seconds <= 0}
            >
                <ModalContent>
                    <ModalHeader>
                        更新日志 & 发电厂
                    </ModalHeader>
                    <ModalBody>
                        {/* 4. 重点：删掉这里的 overflow-scroll，让 ModalBody 接管滚动！ */}
                        <div className={'markdown-body list-disc'}>
                            <Markdown
                                remarkPlugins={[remarkGfm]}
                                rehypePlugins={[rehypeRaw]}
                                components={{
                                    ul: ({node, ...props}) => <ul style={{listStyleType: 'disc', paddingLeft: '2em'}} {...props} />,
                                    "bili-dynamic-card": ({ node, ...props }) => {
                                        return <DynamicCard item={{OID:props.oid}} onClick={() => {
                                            window.open('https://t.bilibili.com/' + props.oid)
                                        }}></DynamicCard>;
                                    },
                                    "bv-player": ({ node, ...props }) => {
                                        return <BVPlayer bv={props.bv}></BVPlayer>
                                    }
                                }}
                            >
                                {props.content}
                            </Markdown>
                        </div>
                    </ModalBody>
                    <ModalFooter>
                        <Button
                            onClick={props.onClose}
                            isDisabled={seconds > 0}
                            color={seconds > 0 ? 'default' : 'primary'}
                        >
                            Close {seconds > 0 ? `${seconds} Seconds` : ''}
                        </Button>
                    </ModalFooter>
                </ModalContent>
            </Modal>
        </div>
    );
}

export default NoticeDialog;