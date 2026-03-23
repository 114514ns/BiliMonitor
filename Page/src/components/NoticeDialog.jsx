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
        <div className={'h-[66vh]'} data-nosnippet>
            <Modal onClose={seconds <=0 && props.onClose} isOpen={true} size={isMobile()?'full':'2xl'} className={'max-h-2/3 sm:h-[75vh]'} scrollBehavior={'inside'} backdrop={'blur'} isDismissable={seconds<=0}>
                <ModalContent>
                    <ModalHeader>
                        About & Notice

                    </ModalHeader>
                    <ModalBody>
                        <div className={'markdown-body  overflow-scroll list-disc' }>
                            <Markdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeRaw]}
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
                            >{props.content}</Markdown>
                        </div>

                    </ModalBody>
                    <ModalFooter>
                        <Button onClick={props.onClose} isDisabled={seconds >0} color={seconds>0?'':'primary'}>Close {seconds>0? `${seconds} Seconds`:''}</Button>
                    </ModalFooter>
                </ModalContent>
            </Modal>
        </div>
    );
}

export default NoticeDialog;