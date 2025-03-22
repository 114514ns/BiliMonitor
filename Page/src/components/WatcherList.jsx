import React from "react";
import { Avatar, Card, CardBody, Chip } from "@heroui/react";
import { useVirtualizer } from "@tanstack/react-virtual";
import { CheckIcon } from "../pages/ChatPage";

function WatcherList(props) {
    const parentRef = React.useRef(null);
    const rowVirtualizer = useVirtualizer({
        count: props.list.length,
        getScrollElement: () => parentRef.current,
        estimateSize: () => 80,
        overscan: 3,
    });

    return (
        <div>
            <div
                ref={parentRef}
                style={{
                    height: "530px",
                    overflow: "auto",
                    position: "relative",
                }}
            >
                <div
                    style={{
                        height: `${rowVirtualizer.getTotalSize()}px`,
                        position: "relative",
                        width: "100%",
                    }}
                >
                    {rowVirtualizer.getVirtualItems().map((virtualRow) => {
                        const item = props.list[virtualRow.index];

                        return (
                            <Card
                                key={item.uid}
                                ref={virtualRow.measureElement}
                                style={{
                                    position: "absolute",
                                    top: 0,
                                    left: 0,
                                    width: "100%",
                                    height: `${virtualRow.size}px`,
                                    transform: `translateY(${virtualRow.start}px)`,
                                }}
                                radius={'none'}
                                shadow={'none'}
                                isHoverable
                            >
                                <CardBody>
                                    <div style={{ display: "flex", alignItems: "center" }}>
                                        <Avatar src={item.face} />
                                        <div style={{ marginLeft: "10px" }}>
                                            <p>{item.name}</p>
                                            {item.medal_info.medal_name && (
                                                <Chip
                                                    startContent={<CheckIcon size={18} />}
                                                    variant="faded"
                                                    style={{
                                                        background: item.medal_info.Color,
                                                        color: "white",
                                                    }}
                                                >
                                                    {item.medal_info.medal_name}
                                                    <span className="text-xs font-bold px-2 py-0.5 rounded-full">
                                                        {item.medal_info.level}
                                                    </span>
                                                </Chip>
                                            )}
                                        </div>
                                    </div>
                                </CardBody>
                            </Card>
                        );
                    })}
                </div>
            </div>
        </div>
    );
}

export default React.memo(WatcherList);