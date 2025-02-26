import {Avatar, Badge, Card, CardBody, CardFooter, CardHeader, Image} from "@heroui/react";


function LiveCard({liveData}) {

    const host = location.hostname;

    const port = debug ? 8080 : location.port;

    const protocol = location.protocol.replace(":", "")
    const {Live, UName, UID, Area, Title, Face, Cover} = liveData;
    const cover = `${protocol}://${host}:${port}/proxy?url=${Cover}`
    return (
        <div>
            <Card style={{
                width: 300,
                marginRight: '20px',
                margin: '15px',
                //backgroundImage: `url(${cover})`,
                //backgroundSize: 'cover',
                //backgroundColor: 'rgba(0, 0, 0, 0.5)', // 透明度 50% 的黑色遮罩
                //backgroundBlendMode: 'overlay' // 让颜色与背景融合
            }}
            >

                <CardHeader className="flex items-center gap-3">
                    <Badge color={Live ? "success" : "default"} content="">
                        <Avatar src={`${protocol}://${host}:${port}/proxy?url=${Face}`}/>
                    </Badge>

                    <div>
                        <h4 className="font-semibold">{UName}</h4>
                        <p className="text-gray-500">{Title}</p>
                    </div>
                </CardHeader>
                <CardBody style={{overflow: 'hidden'}}>
                    <Image
                        removeWrapper
                        alt="Card background"
                        className="z-0 w-full h-full object-cover"
                        src={cover}
                        isBlurred
                        isZoomed
                    />
                </CardBody>
                <CardFooter className="flex justify-between">
                    <span className="text-gray-500 font-semibold">Area: {Area}</span>
                    <span
                        className="text-blue-500 font-semibold cursor-pointer"
                        onClick={() => window.open(`https://space.bilibili.com/${UID}`)}
                    >
                    UID: {UID}
                </span>
                </CardFooter>
            </Card>
        </div>
    )
}

export default LiveCard;