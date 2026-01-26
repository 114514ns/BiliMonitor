import React, {useEffect, useState} from 'react';
import {Button, Modal, ModalBody, ModalContent, ModalFooter, ModalHeader} from "@heroui/react";
import Markdown from "react-markdown";
import axios from "axios";
import remarkGfm from 'remark-gfm'
import "github-markdown-css/github-markdown-light.css";
import rehypeRaw from "rehype-raw";

function NoticeDialog(props) {


    const [seconds,setSeconds] = useState(5);

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

    return (
        <div className={'max-h-2/3'} data-nosnippet>
            <Modal onClose={props.onClose} isOpen={true} size={'lg'} className={'max-h-1/2 sm:max-h-2/3'} scrollBehavior={'inside'} backdrop={'blur'} isDismissable={seconds<=0}>
                <ModalContent>
                    <ModalHeader>
                        About & Notice

                    </ModalHeader>
                    <ModalBody>
                        <div className={'markdown-body  max-h-1/3 overflow-scroll list-disc' }>
                            <Markdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeRaw]}
                                      components={{
                                          ul: ({node, ...props}) => <ul style={{listStyleType: 'disc', paddingLeft: '2em'}} {...props} />
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