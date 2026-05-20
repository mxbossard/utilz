package ztring

import "strings"

const (
	loremIpsum = `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed non risus. Suspendisse lectus tortor, dignissim sit amet, adipiscing nec, ultricies sed, dolor. Cras elementum ultrices diam. Maecenas ligula massa, varius a, semper congue, euismod non, mi. Proin porttitor, orci nec nonummy molestie, enim est eleifend mi, non fermentum diam nisl sit amet erat. Duis semper. Duis arcu massa, scelerisque vitae, consequat in, pretium a, enim. Pellentesque congue. Ut in risus volutpat libero pharetra tempor. Cras vestibulum bibendum augue. Praesent egestas leo in pede. Praesent blandit odio eu enim. Pellentesque sed dui ut augue blandit sodales. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Aliquam nibh. Mauris ac mauris sed pede pellentesque fermentum. Maecenas adipiscing ante non diam sodales hendrerit.
Ut velit mauris, egestas sed, gravida nec, ornare ut, mi. Aenean ut orci vel massa suscipit pulvinar. Nulla sollicitudin. Fusce varius, ligula non tempus aliquam, nunc turpis ullamcorper nibh, in tempus sapien eros vitae ligula. Pellentesque rhoncus nunc et augue. Integer id felis. Curabitur aliquet pellentesque diam. Integer quis metus vitae elit lobortis egestas. Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Morbi vel erat non mauris convallis vehicula. Nulla et sapien. Integer tortor tellus, aliquam faucibus, convallis id, congue eu, quam. Mauris ullamcorper felis vitae erat. Proin feugiat, augue non elementum posuere, metus purus iaculis lectus, et tristique ligula justo vitae magna.
Aliquam convallis sollicitudin purus. Praesent aliquam, enim at fermentum mollis, ligula massa adipiscing nisl, ac euismod nibh nisl eu lectus. Fusce vulputate sem at sapien. Vivamus leo. Aliquam euismod libero eu enim. Nulla nec felis sed leo placerat imperdiet. Aenean suscipit nulla in justo. Suspendisse cursus rutrum augue. Nulla tincidunt tincidunt mi. Curabitur iaculis, lorem vel rhoncus faucibus, felis magna fermentum augue, et ultricies lacus lorem varius purus. Curabitur eu amet.`
	mdLoremIpsum = `# Ignes odioque rigidas conterruit

## Latus omnibus ne

Lorem markdownum inanis. Haec carpitur, ruinas pedum cessantem artus referebam
postquam canna. Et est saligno duce?

    var android_post = eps_vaporware + 1 * lanMemory + vertical_file(excel,
            pci_ppl);
    if (18) {
        twainDvdFios.trackbackSignature -= 1;
    } else {
        bounce.rightLeafMyspace.layoutCopyPoint(machine + snippet_flood,
                whitePpmNybble / rgb_cloud_drive);
        burnUnfriendParity.oem += 5;
        language.whitelist_heuristic = 4;
    }
    var day_smishing = 2;

Fingit erant *adversum* inhaesuro viginti emissumque non rupes carmine
'constant_pipeline' colorem. Sic quae exululat. Cura posse efficient abstulit
cernit quidem! Verbis nunc odissem lacrimavit alipedis pectus; femina beatum in
ferarum 'hypermediaSoftware' caelestibus meus in terna radice? Non est cruorem
fatali!

> Est tua similis et superi gurgite sub corripe fecissem agger modo sint est,
> modo. Sequente ubi litem amans 'eLatencyLocalhost' vocabula splendidior illo?
> Sine sit Tmolo perque [pecudes nostro](#vitamque-cum-quid-corpora), ausa, quae
> quoque, ales fit quaerite.

## Vitamque cum quid corpora

### Volucri et densum nemorum

Pectus uritur, cura fulget color remotos; *herculis et gramen* annis medicamina
viribus mortalia naides cetera. *Meliora flamma* supplex *cultum mollierat vias*
in tecta, frequentat **iura flendo certare**. Monitis dira. Pendere membra,
quaedam, **corpora**, carmine venter ipsa. Ore habe vultus *robur* nam; et agros
in quam, misistis pari?

> Tollit annis? Amor lingua, et cuius: dea hamatis pueri, [mentita
> adde](#ignes-odioque-rigidas-conterruit), mixtam abiturus cesserunt solebat,
> Qui surgere? Exclamant corporis fecit. Sinuessa curvum, tempore 'drive_bar'
> sit: iam [si poteram](#et-laeva-urit-tum) accensae tellus abscessisse inquit.
> Cui inpia 'leaderboardBluetooth' si duos functus.

### Domus erit sui caret

Ore oblita, coepit in et tardae neque gemitus mihi ultro, quotiens huc deum huc?
Fatebor nec **tu patens** gramen. Redeant fuit sic cum non, vis sentit vade
perstat crura neutrumque origine Pelori; thalamos.

Nimia dona terras pondere et unum **temptant armis**, lucem tibi Circaea quoque,
nemorum? Condidit et quod; cum sit cum nec erat gratia membra Eumenides primas
adrides fatigat, quo duos. Est utrumque, Athamanteosque lege, omnem, vel
'unit_trash' capacibus tenebrae **solvit versuta**. Quae licet feroces gravi est
forma **parentis somnus** fama amem meo satiaque, vidit.

### Et laeva urit tum

[Nec tecti](#et-laeva-urit-tum). Sub alimentaque quicquam secreta facinus
fatorum voces. Insilit ille titulum similis capit, ossa in Iuli adamanta dei
sermo!
`
	schnapsum = "Lorem Elsass ipsum dui pellentesque mänele Kabinetpapier tchao bissame Carola und vielmols, wie dignissim rhoncus tristique purus Yo dû. sed mollis baeckeoffe libero, ac Gal. varius Huguette ftomi! suspendisse ornare réchime DNA, kougelhopf Chulien flammekueche hopla turpis, messti de Bischheim wurscht Miss Dahlias Pellentesque Salut bisamme commodo Strasbourg schpeck elit s'guelt dolor sit senectus blottkopf, leo Coopé de Truchtersheim ch'ai rossbolla ornare Oberschaeffolsheim condimentum tellus chambon aliquam so  schnaps leo eget munster gal amet id, yeuh. gravida in, semper morbi turpis salu hoplageiss quam. Hans Chulia Roberstau Mauris bissame hopla ac Salu bissame barapli nüdle libero, bredele tellus kuglopf eleifend Wurschtsalad knepfle placerat leverwurscht météor habitant nullam ante elementum vulputate rucksack non sit amet, Heineken kartoffelsalad Verdammi Spätzle geïz et picon bière lotto-owe Morbi geht's risus, merci vielmols schneck auctor, gewurztraminer Oberschaeffolsheim sed sit mamsell hopla non Pfourtz ! hopla knack Richard Schirmeck jetz gehts los sagittis amet libero. Racing. id adipiscing Christkindelsmärik lacus consectetur Gal ! ullamcorper porta hop quam,"

	saganIpsum = "Ship of the imagination the sky calls to us great turbulent clouds star stuff harvesting star light tingling of the spine extraordinary claims require extraordinary evidence. Invent the universe courage of our questions with pretty stories for which there's little good evidence of brilliant syntheses Jean-François Champollion white dwarf? Concept of the number one something incredible is waiting to be known descended from astronomers intelligent beings not a sunrise but a galaxyrise permanence of the stars and billions upon billions upon billions upon billions upon billions upon billions upon billions."
)

func slice(s string) []string {
	return strings.Split(s, " ")
}

func words(s string, n int) string {
	words := slice(s)
	sb := []string{}
	for i := range n {
		sb = append(sb, words[i%len(words)])
	}
	return strings.Join(sb, " ")
}

func length(s string, l int) string {
	sb := strings.Builder{}
	p := len(s)
	for i := 0; i < l; i += l {
		c := min(l-i, p)
		sb.WriteString(s[i : i+c])
	}
	return sb.String()
}

func LoremIpsum() string {
	return loremIpsum
}

func LoremIpsumSplitedWords() []string {
	return slice(loremIpsum)
}

func LoremIpsumWords(n int) string {
	return words(loremIpsum, n)
}

func LoremIpsumLength(l int) string {
	return length(loremIpsum, l)
}

func MdLoremIpsum() string {
	return mdLoremIpsum
}

func MdLoremIpsumSplitedWords() []string {
	return slice(mdLoremIpsum)
}

func MdLoremIpsumWords(n int) string {
	return words(mdLoremIpsum, n)
}

func MdLoremIpsumLength(l int) string {
	return length(mdLoremIpsum, l)
}

func Schnapsum() string {
	return schnapsum
}

func SchnapsumSplitedWords() []string {
	return slice(schnapsum)
}

func SchnapsumWords(n int) string {
	return words(schnapsum, n)
}

func SchnapsumLength(l int) string {
	return length(schnapsum, l)
}

func SaganIpsum() string {
	return saganIpsum
}

func SaganIpsumSplitedWords() []string {
	return slice(saganIpsum)
}

func SaganIpsumWords(n int) string {
	return words(saganIpsum, n)
}

func SaganIpsumLength(l int) string {
	return length(saganIpsum, l)
}

