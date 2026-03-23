import React, {useEffect, useState} from 'react';
import {
    Modal,
    ModalContent,
    ModalHeader,
    ModalBody,
    ModalFooter,
    Button,
    useDisclosure, Input, addToast, ToastProvider,
} from "@heroui/react";

import axios from "axios";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeRaw from "rehype-raw";
import rehypeSanitize, {defaultSchema} from 'rehype-sanitize'
import {useTheme} from "next-themes";
import DynamicCard from "./DynamicCard";
import BVPlayer from "./BVPlayer";

function DelectIcon() {
    return (
        <svg xmlns="http://www.w3.org/2000/svg" className="icon" viewBox="0 0 1024 1024" width={24} height={24}>
            <path
                d="M360 184h-8c4.4 0 8-3.6 8-8v8h304v-8c0 4.4 3.6 8 8 8h-8v72h72v-80c0-35.3-28.7-64-64-64H352c-35.3 0-64 28.7-64 64v80h72v-72zm504 72H160c-17.7 0-32 14.3-32 32v32c0 4.4 3.6 8 8 8h60.4l24.7 523c1.6 34.1 29.8 61 63.9 61h454c34.2 0 62.3-26.8 63.9-61l24.7-523H888c4.4 0 8-3.6 8-8v-32c0-17.7-14.3-32-32-32zM731.3 840H292.7l-24.2-512h487l-24.2 512z"/>
        </svg>
    )
}
function isNumber(str) {
    return !isNaN(Number(str)) && str.trim() !== '';
}

function CommentForm(props) {
    const sanitizeSchema = {
        ...defaultSchema,
        tagNames:[...(defaultSchema.tagNames || []), 'audio', 'video', 'source','bv-player'],
        attributes: {
            ...defaultSchema.attributes,
            audio:['src', 'controls', 'loop', 'muted', 'autoplay'],
            video:['src', 'controls', 'width', 'height', 'loop', 'muted', 'autoplay', 'poster'],
            source: ['src', 'type'],
            'bv-player':['bv'],
        }
    };


    const [text, setText] = useState('');

    const [page, setPage] = useState(1);

    const [pageSize, setPageSize] = useState(500);

    const [list, setList] = useState([])

    const [total, setTotal] = useState(0);

    const refresh = () => {
        axios.get(`/api/comments/list?session=${localStorage.getItem('session')}&page=${page}&size=${pageSize}`).then((res) => {
            setList(res.data.data??[]);
            setTotal(res.data.total_pages);
        })
    }

    useEffect(() => {
        refresh();
    },[page,pageSize,props.isOpen])

    useEffect(() => {
        const ref = setInterval(() => {
            refresh();
        },1000*90)
        return () => {
            clearInterval(ref);
        }
    },[])
    const { theme } = useTheme();

    const resolvedTheme = theme === "dark" ? "dark" : "light";

    return (
        <div className={'max-h-[60vh]'}>
            <ToastProvider placement={'top-right'}/>
            <Modal isOpen={props.isOpen} onOpenChange={props.onChange} scrollBehavior={'inside'}>
                <ModalContent>
                    <ModalHeader className="flex flex-col gap-1">Comments</ModalHeader>
                    <ModalBody>
                        <div className={'overflow-y-scroll max-h-[50vh]'}>
                            {list.map((item,i) => {
                                return (
                                    <div className={'flex flex-row items-center markdown-body'  }>
                                        {item.Text !== '' && <p>{item.DisplayName}: {  <Markdown remarkPlugins={[remarkGfm]}    rehypePlugins={[rehypeRaw,[rehypeSanitize, sanitizeSchema]]}
                                                                                                 components={{
                                                                                                     a: ({node, ...props}) => <a className="text-blue-500 hover:text-blue-700 hover:underline" target="_blank" rel="noreferrer" {...props} />,
                                                                                                     p: 'span',
                                                                                                     img: ({node, ...props}) => <img {...props} className="inline-block" />,
                                                                                                     audio: ({node, ...props}) => <audio {...props} className="inline-block h-8" />,
                                                                                                     ul: ({node, ...props}) => <ul style={{listStyleType: 'disc', paddingLeft: '2em'}} {...props} />,
                                                                                                     "bili-dynamic-card": ({ node, ...props }) => {
                                                                                                         return <DynamicCard OID={props.OID}></DynamicCard>;
                                                                                                     },
                                                                                                     "bv-player": ({ node, ...props }) => {
                                                                                                         return <BVPlayer bv={props.bv}></BVPlayer>
                                                                                                     }
                                                                                                 }}
                                        >{item.Text}</Markdown>}</p>}
                                        {item.Session !== '' && <Button isIconOnly startContent={<DelectIcon/>} onClick={() => {
                                            axios.post('/api/comments/delete',
                                                new URLSearchParams({
                                                    id: item.ID,
                                                    session: localStorage.getItem('session')
                                                })
                                            ).then(() => {
                                                refresh();
                                            })

                                        }} className={'ml-4'}/> }
                                    </div>
                                )
                            })}

                        </div>
                        <div className={'flex flex-row items-center'}>
                            <Input label={'Texts'} onValueChange={text => setText(text)} value={text} />
                            <Button className={'ml-2'} onClick={() => {

                                if (isNumber(text) || isNumber(text.replace('UID:',''))) {
                                    addToast({
                                        title: 'Alert',
                                        description: 'DONT SPAM（不要发送无意义内容）',
                                        color: 'danger',
                                    })
                                } else {
                                    axios.post('/api/comments/send',
                                        new URLSearchParams({
                                            text: text,
                                            session: localStorage.getItem('session')
                                        })
                                    ).then(() => {
                                        refresh();
                                    })
                                }
                                setText('')
                            }}>Send</Button>
                        </div>
                    </ModalBody>
                    <ModalFooter>
                        <Button color="primary" onPress={props.onClose}>
                            Close
                        </Button>
                    </ModalFooter>
                </ModalContent>
            </Modal>
        </div>
    );
}

export default CommentForm;