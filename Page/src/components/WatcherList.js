import React from 'react';
import {Avatar, Card, CardBody} from "@heroui/react";
import {useVirtualizer} from "@tanstack/react-virtual";

function WatcherList(props) {
    const ref = React.useRef(null);
    const rowVirtualizer = useVirtualizer({
        count: message.length,
        getScrollElement: () => ref.current, // 绑定滚动容器
        estimateSize: () => 160, // 预估行高
        overscan: 30
    });
    return (
        <div>
            <div style={{ height: `${rowVirtualizer.getTotalSize()}px`, position: "relative" }}>
                {rowVirtualizer.getVirtualItems().map(virtualRow => {
                    const item = message[virtualRow.index];
                    return (
                        <Card>
                            <CardBody>
                                <Avatar src={item.face}>

                                </Avatar>
                            </CardBody>
                        </Card>
                    )
                })}
            </div>
        </div>
    );
}

export default WatcherList;