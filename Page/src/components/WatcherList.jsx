import React, {useEffect} from "react";
import { Avatar, Card, CardBody, Chip } from "@heroui/react";
import { useVirtualizer } from "@tanstack/react-virtual";
import { CheckIcon } from "../pages/ChatPage";
import axios from "axios";

function WatcherList(props) {
    const parentRef = React.useRef(null);
    const [list, setList] = React.useState([]);
    const host = location.hostname;

    const port =  location.port;

    const protocol = location.protocol.replace(":", "")
    const rowVirtualizer = useVirtualizer({
        count: list.length,
        getScrollElement: () => parentRef.current,
        estimateSize: () => 80,
        overscan: 3,
    });
    const refresh = () => {
        axios.get(`${protocol}://${host}:${port}/api/monitor/${props.room}`).then((response) => {
            if (props.type=="guard") {
                setList(response.data.live.GuardList);
            } else {
                setList(response.data.live.OnlineWatcher);
            }
        })
    }
    const getFrame = (level) => {
        if (level == 2) {
            return `${protocol}://${host}:${port}/proxy?url=https://i0.hdslb.com/bfs/live/3b46129e796df42ec7356fcba77c8a79d47db682.png@50w_50h.webp`
        }
        if (level == 3) {
            return `${protocol}://${host}:${port}/proxy?url=https://i0.hdslb.com/bfs/live/3bc68207932eabf980cd6a0dd09f4d24f9cc26da.png@50w_50h.webp`;
        }
        if (level == 1) {
            return `${protocol}://${host}:${port}/proxy?url=https://i0.hdslb.com/bfs/live/a454275dea465ac15a03f121f0d7edaf96e30bcf.png@50w_50h.webp`;
        }
        return ""

    }

    useEffect(() => {
        refresh();
        const interval = setInterval(() => {
            refresh();
        }, 5000);

        return () => {
            clearInterval(interval);
        }
    }, [props.room]);

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
                        const item = list[virtualRow.index];

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
                                        <div
                                            style={{
                                                position: 'relative', // 让子元素可以用 `absolute`
                                                width: '50px',
                                                height: '50px'
                                            }}
                                        >
                                            <Avatar
                                                src={item.face}
                                                style={{
                                                    backgroundSize: 'cover',
                                                    backgroundPosition: 'center',
                                                    width: '100%',
                                                    height: '100%',
                                                    borderRadius: '50%',
                                                    position: 'absolute',
                                                    top: 0,
                                                    left: 0,
                                                    zIndex: 1 // 头像层级较低
                                                }}
                                            />

                                            <div
                                                style={{
                                                    backgroundImage: `url(${getFrame(item.guard_level)})`,
                                                    backgroundSize: 'cover',
                                                    backgroundPosition: 'center',
                                                    width: '100%',
                                                    height: '100%',
                                                    position: 'absolute',
                                                    top: 0,
                                                    left: 0,
                                                    zIndex: 2 // 头像框在上层
                                                }}
                                            />
                                        </div>


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
const CacheAvatar = React.memo(({ src }) => {
    return <Avatar src={src}/>;
});
export default React.memo(WatcherList);