import React, {useEffect} from 'react';
import {Button, Modal, ModalBody, ModalContent, ModalFooter, ModalHeader, useDisclosure} from "@heroui/react";
import {useLocation} from "react-router-dom";
import Markdown from "react-markdown";
import rehypeRaw from "rehype-raw";
import remarkGfm from "remark-gfm";
import axios from "axios";
import DynamicCard from "./DynamicCard";

function DocsDialog(props) {

    const loc = useLocation();

    const [content,setContent] = React.useState('');

    useEffect(() => {
        var p = loc.pathname
        var fName = ''
        if (p === '/') {
            fName = 'index.md'
        }
        if (p === '/list') {
            fName = 'list.md'
        }

        if (p === '/raw') {
            fName = 'raw.md'
        }

        if (p === '/pk') {
            fName = 'pk.md'
        }

        if (p === '/traces') {
            fName = 'traces.md'
        }

        if (p === '/reactions') {
            fName = 'reactions.md'
        }

        if (p === '/relation') {
            fName = 'relation.md'
        }

        if (p === '/fans') {
            fName = 'fans.md'
        }
        if (p === '/feeds') {
            fName = 'dynamics.md'
        }
        if (p === '/highlight') {
            fName = 'highlight.md'
        }
        if (fName !== '') {
            axios.get(`/docs/${fName}`).then(res => {
                setContent(res.data)
            })
        } else {
            setContent('还没有内容......')
        }
    }, [loc.pathname]);
    return (
        <div className={'max-h-1/2 sm:h-2/3 '}>
            <Modal isOpen={props.isOpen} onOpenChange={props.onClose} className={'overflow-scroll scrollbar-hide'} >
                <ModalContent>
                    {(onClose) => (
                        <>
                            <ModalHeader className="flex flex-col gap-1">🦌</ModalHeader>
                            <ModalBody className={'markdown-body list-disc'}>
                                <Markdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeRaw]}      components={{
                                    ul: ({node, ...props}) => <ul style={{listStyleType: 'disc', paddingLeft: '2em'}} {...props} />,
                                    "bili-dynamic-card": ({ node, ...props }) => {

// props 中包含了 type 属性以及 children

                                        return <DynamicCard OID={props.OID}></DynamicCard>;

                                    },
                                }}>
                                    {content}
                                </Markdown>
                            </ModalBody>
                            <ModalFooter>
                                <Button color="danger" variant="light" onPress={props.onClose}>
                                    Close
                                </Button>
                            </ModalFooter>
                        </>
                    )}
                </ModalContent>
            </Modal>
        </div>
    );
}

export default DocsDialog;